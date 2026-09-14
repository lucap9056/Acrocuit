using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Storage;

namespace Acrocuit.StoredProcedures;

public static class SpGetDeviceUpstream
{
    public const string Name = "sp_GetDeviceUpstream";

    private const string USER_ID = "@UserId";
    private const string DEVICE_ID = "@DeviceId";
    private const string FOUND = "@Found";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {USER_ID} INT,
            {DEVICE_ID} INT,
            {FOUND} BIT OUTPUT
        AS
        BEGIN
            SET NOCOUNT ON;

            IF NOT EXISTS (
                SELECT 1
                FROM Device d
                INNER JOIN Space s ON s.Id = d.SpaceId
                INNER JOIN SpaceGroupOwner so ON so.SpaceGroupId = s.SpaceGroupId
                WHERE d.Id = {DEVICE_ID} AND so.UserId = {USER_ID}
            )
            BEGIN
                SET {FOUND} = 0;
                SELECT
                    CAST(0 AS INT) AS Id,
                    CAST('' AS NVARCHAR(255)) AS Name,
                    CAST(0 AS INT) AS BreakerGroupId,
                    CAST(0 AS INT) AS SpaceGroupId,
                    CAST(0 AS INT) AS DisplayOrder,
                    CAST(NULL AS INT) AS UpstreamBreaker
                WHERE 1 = 0;
                RETURN;
            END

            SET {FOUND} = 1;

            ;WITH BreakerChain AS (
                SELECT b.Id, b.Name, b.BreakerGroupId, b.SpaceGroupId, b.DisplayOrder, b.UpstreamBreaker
                FROM DeviceBreaker db
                INNER JOIN Breaker b ON b.Id = db.BreakerId
                WHERE db.DeviceId = {DEVICE_ID}

                UNION ALL

                SELECT p.Id, p.Name, p.BreakerGroupId, p.SpaceGroupId, p.DisplayOrder, p.UpstreamBreaker
                FROM Breaker p
                INNER JOIN BreakerChain c ON p.Id = c.UpstreamBreaker
            )
            SELECT DISTINCT Id, Name, BreakerGroupId, SpaceGroupId, DisplayOrder, UpstreamBreaker
            FROM BreakerChain
            ORDER BY Id;
        END
        """;

    public readonly record struct Node(int Id, string Name, int BreakerGroupId, int SpaceGroupId, int DisplayOrder, int? UpstreamBreakerId);
    public readonly record struct Result(bool Found, List<Node> Breakers);

    public static async Task<Result> ExecuteAsync(AppDbContext db, int userId, int deviceId, CancellationToken cancellationToken)
    {
        var breakers = new List<Node>();

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
            command.Transaction = db.Database.CurrentTransaction?.GetDbTransaction();

            var foundParam = new SqlParameter(FOUND, SqlDbType.Bit) { Direction = ParameterDirection.Output };
            command.Parameters.Add(new SqlParameter(USER_ID, userId));
            command.Parameters.Add(new SqlParameter(DEVICE_ID, deviceId));
            command.Parameters.Add(foundParam);

            using (var reader = await command.ExecuteReaderAsync(cancellationToken))
            {
                var idOrdinal = reader.GetOrdinal("Id");
                var nameOrdinal = reader.GetOrdinal("Name");
                var breakerGroupIdOrdinal = reader.GetOrdinal("BreakerGroupId");
                var spaceGroupIdOrdinal = reader.GetOrdinal("SpaceGroupId");
                var displayOrderOrdinal = reader.GetOrdinal("DisplayOrder");
                var upstreamBreakerOrdinal = reader.GetOrdinal("UpstreamBreaker");

                while (await reader.ReadAsync(cancellationToken))
                {
                    breakers.Add(new Node(
                        reader.GetInt32(idOrdinal),
                        reader.GetString(nameOrdinal),
                        reader.GetInt32(breakerGroupIdOrdinal),
                        reader.GetInt32(spaceGroupIdOrdinal),
                        reader.GetInt32(displayOrderOrdinal),
                        reader.IsDBNull(upstreamBreakerOrdinal) ? null : reader.GetInt32(upstreamBreakerOrdinal)));
                }
            }

            return new Result((bool)foundParam.Value, breakers);
        }
        finally
        {
            if (shouldClose)
            {
                await connection.CloseAsync();
            }
        }
    }
}
