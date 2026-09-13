using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public enum SetBreakerUpstreamResult
{
    Success = 0,
    BreakerNotFound = 1,
    SelfReference = 2,
    UpstreamNotFound = 3,
    WouldCreateCycle = 4
}

public static class SpSetBreakerUpstream
{
    public const string Name = "sp_SetBreakerUpstream";

    private const string USER_ID = "@UserId";
    private const string BREAKER_ID = "@BreakerId";
    private const string UPSTREAM_BREAKER_ID = "@UpstreamBreakerId";
    private const string RESULT = "@Result";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {USER_ID} INT,
            {BREAKER_ID} INT,
            {UPSTREAM_BREAKER_ID} INT = NULL,
            {RESULT} INT OUTPUT
        AS
        BEGIN
            SET NOCOUNT ON;

            DECLARE @SpaceGroupId INT;

            SELECT @SpaceGroupId = b.SpaceGroupId
            FROM Breaker b
            INNER JOIN BreakerGroup bg ON bg.Id = b.BreakerGroupId
            INNER JOIN SpaceGroupOwner so ON so.SpaceGroupId = bg.SpaceGroupId
            WHERE b.Id = {BREAKER_ID} AND so.UserId = {USER_ID};

            IF @SpaceGroupId IS NULL
            BEGIN
                SET {RESULT} = 1;
                RETURN;
            END

            IF {UPSTREAM_BREAKER_ID} IS NOT NULL
            BEGIN
                IF {UPSTREAM_BREAKER_ID} = {BREAKER_ID}
                BEGIN
                    SET {RESULT} = 2;
                    RETURN;
                END

                IF NOT EXISTS (SELECT 1 FROM Breaker WHERE Id = {UPSTREAM_BREAKER_ID} AND SpaceGroupId = @SpaceGroupId)
                BEGIN
                    SET {RESULT} = 3;
                    RETURN;
                END

                DECLARE @HasCycle BIT = 0;

                ;WITH UpstreamChain AS (
                    SELECT b.Id, b.UpstreamBreaker
                    FROM Breaker b
                    WHERE b.Id = {UPSTREAM_BREAKER_ID}

                    UNION ALL

                    SELECT p.Id, p.UpstreamBreaker
                    FROM Breaker p
                    INNER JOIN UpstreamChain c ON p.Id = c.UpstreamBreaker
                )
                SELECT @HasCycle = 1
                FROM UpstreamChain
                WHERE Id = {BREAKER_ID};

                IF @HasCycle = 1
                BEGIN
                    SET {RESULT} = 4;
                    RETURN;
                END
            END

            UPDATE Breaker SET UpstreamBreaker = {UPSTREAM_BREAKER_ID} WHERE Id = {BREAKER_ID};
            SET {RESULT} = 0;
        END
        """;

    public static async Task<SetBreakerUpstreamResult> ExecuteAsync(AppDbContext db, int userId, int breakerId, int? upstreamBreakerId, CancellationToken cancellationToken)
    {
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

            var resultParam = new SqlParameter(RESULT, SqlDbType.Int) { Direction = ParameterDirection.Output };
            command.Parameters.Add(new SqlParameter(USER_ID, userId));
            command.Parameters.Add(new SqlParameter(BREAKER_ID, breakerId));
            command.Parameters.Add(new SqlParameter(UPSTREAM_BREAKER_ID, (object?)upstreamBreakerId ?? DBNull.Value));
            command.Parameters.Add(resultParam);

            await command.ExecuteNonQueryAsync(cancellationToken);

            return (SetBreakerUpstreamResult)(int)resultParam.Value;
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
