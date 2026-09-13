using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public static class SpGetBreakerUpstreamChain
{
    public const string Name = "sp_GetBreakerUpstreamChain";

    private const string BREAKER_ID = "@BreakerId";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {BREAKER_ID} INT
        AS
        BEGIN
            SET NOCOUNT ON;

            WITH BreakerChain AS (
                SELECT b.Id, b.Name, b.BreakerGroupId, b.SpaceGroupId, b.DisplayOrder, b.UpstreamBreaker, 0 AS Depth
                FROM Breaker b
                WHERE b.Id = {BREAKER_ID}

                UNION ALL

                SELECT p.Id, p.Name, p.BreakerGroupId, p.SpaceGroupId, p.DisplayOrder, p.UpstreamBreaker, c.Depth + 1
                FROM Breaker p
                INNER JOIN BreakerChain c ON p.Id = c.UpstreamBreaker
            )
            SELECT Id, Name, BreakerGroupId, SpaceGroupId, DisplayOrder, UpstreamBreaker
            FROM BreakerChain
            ORDER BY Depth;
        END
        """;

    public readonly record struct Node(int Id, string Name, int BreakerGroupId, int SpaceGroupId, int DisplayOrder, int? UpstreamBreakerId);

    public static async Task<List<Node>> ExecuteAsync(AppDbContext db, int breakerId, CancellationToken cancellationToken)
    {
        var results = new List<Node>();

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

            using var reader = await command.ExecuteReaderAsync(cancellationToken);
            var idOrdinal = reader.GetOrdinal("Id");
            var nameOrdinal = reader.GetOrdinal("Name");
            var breakerGroupIdOrdinal = reader.GetOrdinal("BreakerGroupId");
            var spaceGroupIdOrdinal = reader.GetOrdinal("SpaceGroupId");
            var displayOrderOrdinal = reader.GetOrdinal("DisplayOrder");
            var upstreamBreakerOrdinal = reader.GetOrdinal("UpstreamBreaker");

            while (await reader.ReadAsync(cancellationToken))
            {
                results.Add(new Node(
                    reader.GetInt32(idOrdinal),
                    reader.GetString(nameOrdinal),
                    reader.GetInt32(breakerGroupIdOrdinal),
                    reader.GetInt32(spaceGroupIdOrdinal),
                    reader.GetInt32(displayOrderOrdinal),
                    reader.IsDBNull(upstreamBreakerOrdinal) ? null : reader.GetInt32(upstreamBreakerOrdinal)));
            }
        }
        finally
        {
            if (shouldClose)
            {
                await connection.CloseAsync();
            }
        }

        return results;
    }
}
