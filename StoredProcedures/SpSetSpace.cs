using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public enum SetSpaceResult
{
    Success = 0,
    SpaceNotFound = 1
}

public static class SpSetSpace
{
    public const string Name = "sp_SetSpace";

    private const string USER_ID = "@UserId";
    private const string SPACE_ID = "@SpaceId";
    private const string SPACE_NAME = "@Name";
    private const string DISPLAY_ORDER = "@DisplayOrder";
    private const string OUT_NAME = "@OutName";
    private const string OUT_SPACE_GROUP_ID = "@OutSpaceGroupId";
    private const string OUT_DISPLAY_ORDER = "@OutDisplayOrder";
    private const string OUT_BACKGROUND_IMAGE_UPDATED_AT = "@OutBackgroundImageUpdatedAt";
    private const string RESULT = "@Result";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {USER_ID} INT,
            {SPACE_ID} INT,
            {SPACE_NAME} NVARCHAR(255) = NULL,
            {DISPLAY_ORDER} INT = NULL,
            {OUT_NAME} NVARCHAR(255) OUTPUT,
            {OUT_SPACE_GROUP_ID} INT OUTPUT,
            {OUT_DISPLAY_ORDER} INT OUTPUT,
            {OUT_BACKGROUND_IMAGE_UPDATED_AT} DATETIME2 OUTPUT,
            {RESULT} INT OUTPUT
        AS
        BEGIN
            SET NOCOUNT ON;
            SET XACT_ABORT ON;
            BEGIN TRANSACTION;

            DECLARE @SpaceGroupId INT, @CurrentOrder INT;

            SELECT @SpaceGroupId = s.SpaceGroupId, @CurrentOrder = s.DisplayOrder
            FROM Space s
            INNER JOIN SpaceGroup sg WITH (UPDLOCK, ROWLOCK) ON sg.Id = s.SpaceGroupId
            INNER JOIN SpaceGroupOwner so ON so.SpaceGroupId = sg.Id
            WHERE s.Id = {SPACE_ID} AND so.UserId = {USER_ID};

            IF @SpaceGroupId IS NULL
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
                    UPDATE Space
                    SET DisplayOrder = CASE WHEN Id = {SPACE_ID} THEN {DISPLAY_ORDER} ELSE @CurrentOrder END
                    WHERE SpaceGroupId = @SpaceGroupId AND (Id = {SPACE_ID} OR DisplayOrder = {DISPLAY_ORDER});
                END
                ELSE IF @Delta > 0
                BEGIN
                    UPDATE Space
                    SET DisplayOrder = CASE WHEN Id = {SPACE_ID} THEN {DISPLAY_ORDER} ELSE DisplayOrder - 1 END
                    WHERE SpaceGroupId = @SpaceGroupId AND DisplayOrder >= @CurrentOrder AND DisplayOrder <= {DISPLAY_ORDER};
                END
                ELSE
                BEGIN
                    UPDATE Space
                    SET DisplayOrder = CASE WHEN Id = {SPACE_ID} THEN {DISPLAY_ORDER} ELSE DisplayOrder + 1 END
                    WHERE SpaceGroupId = @SpaceGroupId AND DisplayOrder >= {DISPLAY_ORDER} AND DisplayOrder <= @CurrentOrder;
                END
            END

            IF {SPACE_NAME} IS NOT NULL
            BEGIN
                UPDATE Space SET Name = {SPACE_NAME} WHERE Id = {SPACE_ID};
            END

            SELECT
                {OUT_NAME} = Name,
                {OUT_SPACE_GROUP_ID} = SpaceGroupId,
                {OUT_DISPLAY_ORDER} = DisplayOrder,
                {OUT_BACKGROUND_IMAGE_UPDATED_AT} = BackgroundImageUpdatedAt
            FROM Space
            WHERE Id = {SPACE_ID};

            SET {RESULT} = 0;

            COMMIT TRANSACTION;
        END
        """;

    public readonly record struct Result(SetSpaceResult Status, string Name, int SpaceGroupId, int DisplayOrder, DateTime BackgroundImageUpdatedAt);

    public static async Task<Result> ExecuteAsync(AppDbContext db, int userId, int spaceId, string? name, int? displayOrder, CancellationToken cancellationToken)
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
            var outSpaceGroupIdParam = new SqlParameter(OUT_SPACE_GROUP_ID, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var outDisplayOrderParam = new SqlParameter(OUT_DISPLAY_ORDER, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var outBackgroundImageUpdatedAtParam = new SqlParameter(OUT_BACKGROUND_IMAGE_UPDATED_AT, SqlDbType.DateTime2) { Direction = ParameterDirection.Output };
            var resultParam = new SqlParameter(RESULT, SqlDbType.Int) { Direction = ParameterDirection.Output };

            command.Parameters.Add(new SqlParameter(USER_ID, userId));
            command.Parameters.Add(new SqlParameter(SPACE_ID, spaceId));
            command.Parameters.Add(new SqlParameter(SPACE_NAME, (object?)name ?? DBNull.Value));
            command.Parameters.Add(new SqlParameter(DISPLAY_ORDER, (object?)displayOrder ?? DBNull.Value));
            command.Parameters.Add(outNameParam);
            command.Parameters.Add(outSpaceGroupIdParam);
            command.Parameters.Add(outDisplayOrderParam);
            command.Parameters.Add(outBackgroundImageUpdatedAtParam);
            command.Parameters.Add(resultParam);

            await command.ExecuteNonQueryAsync(cancellationToken);

            var status = (SetSpaceResult)(int)resultParam.Value;

            return status == SetSpaceResult.Success
                ? new Result(
                    status,
                    (string)outNameParam.Value,
                    (int)outSpaceGroupIdParam.Value,
                    (int)outDisplayOrderParam.Value,
                    (DateTime)outBackgroundImageUpdatedAtParam.Value)
                : new Result(status, string.Empty, 0, 0, default);
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
