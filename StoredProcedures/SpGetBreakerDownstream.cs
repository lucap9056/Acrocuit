using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public static class SpGetBreakerDownstream
{
    public const string Name = "sp_GetBreakerDownstream";

    private const string BREAKER_ID = "@BreakerId";
    private const string USER_ID = "@UserId";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {BREAKER_ID} INT,
            {USER_ID} INT = NULL
        AS
        BEGIN
            SET NOCOUNT ON;

            DECLARE @BreakerChain TABLE (
                Id INT PRIMARY KEY,
                Name NVARCHAR(255) NOT NULL,
                BreakerGroupId INT NOT NULL,
                SpaceGroupId INT NOT NULL,
                DisplayOrder INT NOT NULL,
                UpstreamBreaker INT NULL,
                Depth INT NOT NULL
            );

            IF {USER_ID} IS NULL OR EXISTS (
                SELECT 1
                FROM SpaceGroupOwner so
                INNER JOIN BreakerGroup bg ON bg.SpaceGroupId = so.SpaceGroupId
                INNER JOIN Breaker b ON b.BreakerGroupId = bg.Id
                WHERE b.Id = {BREAKER_ID} AND so.UserId = {USER_ID}
            )
            BEGIN
                ;WITH BreakerChain AS (
                    SELECT b.Id, b.Name, b.BreakerGroupId, b.SpaceGroupId, b.DisplayOrder, b.UpstreamBreaker, 0 AS Depth
                    FROM Breaker b
                    WHERE b.Id = {BREAKER_ID}

                    UNION ALL

                    SELECT ch.Id, ch.Name, ch.BreakerGroupId, ch.SpaceGroupId, ch.DisplayOrder, ch.UpstreamBreaker, c.Depth + 1
                    FROM Breaker ch
                    INNER JOIN BreakerChain c ON ch.UpstreamBreaker = c.Id
                )
                INSERT INTO @BreakerChain (Id, Name, BreakerGroupId, SpaceGroupId, DisplayOrder, UpstreamBreaker, Depth)
                SELECT Id, Name, BreakerGroupId, SpaceGroupId, DisplayOrder, UpstreamBreaker, Depth
                FROM BreakerChain;
            END

            SELECT Id, Name, BreakerGroupId, SpaceGroupId, DisplayOrder, UpstreamBreaker
            FROM @BreakerChain
            ORDER BY Depth;

            SELECT d.Id, d.Name, d.SpaceId, db.BreakerId
            FROM Device d
            INNER JOIN DeviceBreaker db ON db.DeviceId = d.Id
            INNER JOIN @BreakerChain c ON db.BreakerId = c.Id
            ORDER BY db.BreakerId, d.Id;
        END
        """;

    public readonly record struct BreakerNode(int Id, string Name, int BreakerGroupId, int SpaceGroupId, int DisplayOrder, int? UpstreamBreakerId);
    public readonly record struct DeviceNode(int Id, string Name, int SpaceId, int BreakerId);
    public readonly record struct Result(List<BreakerNode> Breakers, List<DeviceNode> Devices);

    public static async Task<Result> ExecuteAsync(AppDbContext db, int breakerId, int? userId, CancellationToken cancellationToken)
    {
        var breakers = new List<BreakerNode>();
        var devices = new List<DeviceNode>();

        var connection = db.Database.GetDbConnection();
        var shouldClose = connection.State != ConnectionState.Open;
        if (shouldClose)
        {
            await connection.OpenAsync(cancellationToken);
        }

        try
        {
            using var command = connection.CreateCommand();
            command.CommandText = Name;
            command.CommandType = CommandType.StoredProcedure;
            command.Parameters.Add(new SqlParameter(BREAKER_ID, breakerId));
            command.Parameters.Add(new SqlParameter(USER_ID, (object?)userId ?? DBNull.Value));

            using var reader = await command.ExecuteReaderAsync(cancellationToken);

            var idOrdinal = reader.GetOrdinal("Id");
            var nameOrdinal = reader.GetOrdinal("Name");
            var breakerGroupIdOrdinal = reader.GetOrdinal("BreakerGroupId");
            var spaceGroupIdOrdinal = reader.GetOrdinal("SpaceGroupId");
            var displayOrderOrdinal = reader.GetOrdinal("DisplayOrder");
            var upstreamBreakerOrdinal = reader.GetOrdinal("UpstreamBreaker");

            while (await reader.ReadAsync(cancellationToken))
            {
                breakers.Add(new BreakerNode(
                    reader.GetInt32(idOrdinal),
                    reader.GetString(nameOrdinal),
                    reader.GetInt32(breakerGroupIdOrdinal),
                    reader.GetInt32(spaceGroupIdOrdinal),
                    reader.GetInt32(displayOrderOrdinal),
                    reader.IsDBNull(upstreamBreakerOrdinal) ? null : reader.GetInt32(upstreamBreakerOrdinal)));
            }

            await reader.NextResultAsync(cancellationToken);

            var deviceIdOrdinal = reader.GetOrdinal("Id");
            var deviceNameOrdinal = reader.GetOrdinal("Name");
            var spaceIdOrdinal = reader.GetOrdinal("SpaceId");
            var breakerIdOrdinal = reader.GetOrdinal("BreakerId");

            while (await reader.ReadAsync(cancellationToken))
            {
                devices.Add(new DeviceNode(
                    reader.GetInt32(deviceIdOrdinal),
                    reader.GetString(deviceNameOrdinal),
                    reader.GetInt32(spaceIdOrdinal),
                    reader.GetInt32(breakerIdOrdinal)));
            }
        }
        finally
        {
            if (shouldClose)
            {
                await connection.CloseAsync();
            }
        }

        return new Result(breakers, devices);
    }
}
