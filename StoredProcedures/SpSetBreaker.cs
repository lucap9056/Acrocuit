using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public enum SetBreakerResult
{
    Success = 0,
    BreakerNotFound = 1
}

public static class SpSetBreaker
{
    public const string Name = "sp_SetBreaker";

    private const string USER_ID = "@UserId";
    private const string BREAKER_ID = "@BreakerId";
    private const string BREAKER_NAME = "@Name";
    private const string DISPLAY_ORDER = "@DisplayOrder";
    private const string OUT_NAME = "@OutName";
    private const string OUT_BREAKER_GROUP_ID = "@OutBreakerGroupId";
    private const string OUT_SPACE_GROUP_ID = "@OutSpaceGroupId";
    private const string OUT_DISPLAY_ORDER = "@OutDisplayOrder";
    private const string OUT_UPSTREAM_BREAKER_ID = "@OutUpstreamBreakerId";
    private const string RESULT = "@Result";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {USER_ID} INT,
            {BREAKER_ID} INT,
            {BREAKER_NAME} NVARCHAR(255) = NULL,
            {DISPLAY_ORDER} INT = NULL,
            {OUT_NAME} NVARCHAR(255) OUTPUT,
            {OUT_BREAKER_GROUP_ID} INT OUTPUT,
            {OUT_SPACE_GROUP_ID} INT OUTPUT,
            {OUT_DISPLAY_ORDER} INT OUTPUT,
            {OUT_UPSTREAM_BREAKER_ID} INT OUTPUT,
            {RESULT} INT OUTPUT
        AS
        BEGIN
            SET NOCOUNT ON;
            SET XACT_ABORT ON;
            BEGIN TRANSACTION;

            DECLARE @BreakerGroupId INT, @CurrentOrder INT;

            SELECT @BreakerGroupId = b.BreakerGroupId, @CurrentOrder = b.DisplayOrder
            FROM Breaker b
            INNER JOIN BreakerGroup bg WITH (UPDLOCK, ROWLOCK) ON bg.Id = b.BreakerGroupId
            INNER JOIN SpaceGroupOwner so ON so.SpaceGroupId = bg.SpaceGroupId
            WHERE b.Id = {BREAKER_ID} AND so.UserId = {USER_ID};

            IF @BreakerGroupId IS NULL
            BEGIN
                ROLLBACK TRANSACTION;
                SET {RESULT} = 1;
                RETURN;
            END

            IF {DISPLAY_ORDER} IS NOT NULL AND {DISPLAY_ORDER} <> @CurrentOrder
            BEGIN
                DECLARE @Delta INT = {DISPLAY_ORDER} - @CurrentOrder;

                IF ABS(@Delta) = 1
                BEGIN
                    UPDATE Breaker
                    SET DisplayOrder = CASE WHEN Id = {BREAKER_ID} THEN {DISPLAY_ORDER} ELSE @CurrentOrder END
                    WHERE BreakerGroupId = @BreakerGroupId AND (Id = {BREAKER_ID} OR DisplayOrder = {DISPLAY_ORDER});
                END
                ELSE IF @Delta > 0
                BEGIN
                    UPDATE Breaker
                    SET DisplayOrder = CASE WHEN Id = {BREAKER_ID} THEN {DISPLAY_ORDER} ELSE DisplayOrder - 1 END
                    WHERE BreakerGroupId = @BreakerGroupId AND DisplayOrder >= @CurrentOrder AND DisplayOrder <= {DISPLAY_ORDER};
                END
                ELSE
                BEGIN
                    UPDATE Breaker
                    SET DisplayOrder = CASE WHEN Id = {BREAKER_ID} THEN {DISPLAY_ORDER} ELSE DisplayOrder + 1 END
                    WHERE BreakerGroupId = @BreakerGroupId AND DisplayOrder >= {DISPLAY_ORDER} AND DisplayOrder <= @CurrentOrder;
                END
            END

            IF {BREAKER_NAME} IS NOT NULL
            BEGIN
                UPDATE Breaker SET Name = {BREAKER_NAME} WHERE Id = {BREAKER_ID};
            END

            SELECT
                {OUT_NAME} = Name,
                {OUT_BREAKER_GROUP_ID} = BreakerGroupId,
                {OUT_SPACE_GROUP_ID} = SpaceGroupId,
                {OUT_DISPLAY_ORDER} = DisplayOrder,
                {OUT_UPSTREAM_BREAKER_ID} = UpstreamBreaker
            FROM Breaker
            WHERE Id = {BREAKER_ID};

            SET {RESULT} = 0;

            COMMIT TRANSACTION;
        END
        """;

    public readonly record struct Result(SetBreakerResult Status, string Name, int BreakerGroupId, int SpaceGroupId, int DisplayOrder, int? UpstreamBreakerId);

    public static async Task<Result> ExecuteAsync(AppDbContext db, int userId, int breakerId, string? name, int? displayOrder, CancellationToken cancellationToken)
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

            var outNameParam = new SqlParameter(OUT_NAME, SqlDbType.NVarChar, 255) { Direction = ParameterDirection.Output };
            var outBreakerGroupIdParam = new SqlParameter(OUT_BREAKER_GROUP_ID, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var outSpaceGroupIdParam = new SqlParameter(OUT_SPACE_GROUP_ID, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var outDisplayOrderParam = new SqlParameter(OUT_DISPLAY_ORDER, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var outUpstreamBreakerIdParam = new SqlParameter(OUT_UPSTREAM_BREAKER_ID, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var resultParam = new SqlParameter(RESULT, SqlDbType.Int) { Direction = ParameterDirection.Output };

            command.Parameters.Add(new SqlParameter(USER_ID, userId));
            command.Parameters.Add(new SqlParameter(BREAKER_ID, breakerId));
            command.Parameters.Add(new SqlParameter(BREAKER_NAME, (object?)name ?? DBNull.Value));
            command.Parameters.Add(new SqlParameter(DISPLAY_ORDER, (object?)displayOrder ?? DBNull.Value));
            command.Parameters.Add(outNameParam);
            command.Parameters.Add(outBreakerGroupIdParam);
            command.Parameters.Add(outSpaceGroupIdParam);
            command.Parameters.Add(outDisplayOrderParam);
            command.Parameters.Add(outUpstreamBreakerIdParam);
            command.Parameters.Add(resultParam);

            await command.ExecuteNonQueryAsync(cancellationToken);

            var status = (SetBreakerResult)(int)resultParam.Value;

            return status == SetBreakerResult.Success
                ? new Result(
                    status,
                    (string)outNameParam.Value,
                    (int)outBreakerGroupIdParam.Value,
                    (int)outSpaceGroupIdParam.Value,
                    (int)outDisplayOrderParam.Value,
                    outUpstreamBreakerIdParam.Value is DBNull ? null : (int)outUpstreamBreakerIdParam.Value)
                : new Result(status, string.Empty, 0, 0, 0, null);
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
