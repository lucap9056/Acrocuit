using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public enum AddBreakerResult
{
    Success = 0,
    BreakerGroupNotFound = 1,
    UpstreamNotFound = 2
}

public static class SpAddBreaker
{
    public const string Name = "sp_AddBreaker";

    private const string USER_ID = "@UserId";
    private const string BREAKER_GROUP_ID = "@BreakerGroupId";
    private const string BREAKER_NAME = "@Name";
    private const string UPSTREAM_BREAKER_ID = "@UpstreamBreakerId";
    private const string NEW_BREAKER_ID = "@NewBreakerId";
    private const string SPACE_GROUP_ID = "@SpaceGroupId";
    private const string DISPLAY_ORDER = "@DisplayOrder";
    private const string RESULT = "@Result";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {USER_ID} INT,
            {BREAKER_GROUP_ID} INT,
            {BREAKER_NAME} NVARCHAR(255),
            {UPSTREAM_BREAKER_ID} INT = NULL,
            {NEW_BREAKER_ID} INT OUTPUT,
            {SPACE_GROUP_ID} INT OUTPUT,
            {DISPLAY_ORDER} INT OUTPUT,
            {RESULT} INT OUTPUT
        AS
        BEGIN
            SET NOCOUNT ON;
            SET XACT_ABORT ON;
            BEGIN TRANSACTION;

            SELECT {SPACE_GROUP_ID} = bg.SpaceGroupId
            FROM BreakerGroup bg WITH (UPDLOCK, ROWLOCK)
            INNER JOIN SpaceGroupOwner so ON so.SpaceGroupId = bg.SpaceGroupId
            WHERE bg.Id = {BREAKER_GROUP_ID} AND so.UserId = {USER_ID};

            IF {SPACE_GROUP_ID} IS NULL
            BEGIN
                ROLLBACK TRANSACTION;
                SET {RESULT} = 1;
                RETURN;
            END

            IF {UPSTREAM_BREAKER_ID} IS NOT NULL AND NOT EXISTS (
                SELECT 1 FROM Breaker WHERE Id = {UPSTREAM_BREAKER_ID} AND SpaceGroupId = {SPACE_GROUP_ID}
            )
            BEGIN
                ROLLBACK TRANSACTION;
                SET {RESULT} = 2;
                RETURN;
            END

            SELECT {DISPLAY_ORDER} = ISNULL(MAX(DisplayOrder), 0) + 1
            FROM Breaker
            WHERE BreakerGroupId = {BREAKER_GROUP_ID};

            INSERT INTO Breaker (Name, BreakerGroupId, SpaceGroupId, DisplayOrder, UpstreamBreaker)
            VALUES ({BREAKER_NAME}, {BREAKER_GROUP_ID}, {SPACE_GROUP_ID}, {DISPLAY_ORDER}, {UPSTREAM_BREAKER_ID});

            SET {NEW_BREAKER_ID} = CAST(SCOPE_IDENTITY() AS INT);
            SET {RESULT} = 0;

            COMMIT TRANSACTION;
        END
        """;

    public readonly record struct Result(AddBreakerResult Status, int BreakerId, int SpaceGroupId, int DisplayOrder);

    public static async Task<Result> ExecuteAsync(AppDbContext db, int userId, int breakerGroupId, string name, int? upstreamBreakerId, CancellationToken cancellationToken)
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

            var newBreakerIdParam = new SqlParameter(NEW_BREAKER_ID, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var spaceGroupIdParam = new SqlParameter(SPACE_GROUP_ID, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var displayOrderParam = new SqlParameter(DISPLAY_ORDER, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var resultParam = new SqlParameter(RESULT, SqlDbType.Int) { Direction = ParameterDirection.Output };

            command.Parameters.Add(new SqlParameter(USER_ID, userId));
            command.Parameters.Add(new SqlParameter(BREAKER_GROUP_ID, breakerGroupId));
            command.Parameters.Add(new SqlParameter(BREAKER_NAME, name));
            command.Parameters.Add(new SqlParameter(UPSTREAM_BREAKER_ID, (object?)upstreamBreakerId ?? DBNull.Value));
            command.Parameters.Add(newBreakerIdParam);
            command.Parameters.Add(spaceGroupIdParam);
            command.Parameters.Add(displayOrderParam);
            command.Parameters.Add(resultParam);

            await command.ExecuteNonQueryAsync(cancellationToken);

            var status = (AddBreakerResult)(int)resultParam.Value;

            return status == AddBreakerResult.Success
                ? new Result(status, (int)newBreakerIdParam.Value, (int)spaceGroupIdParam.Value, (int)displayOrderParam.Value)
                : new Result(status, 0, 0, 0);
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
