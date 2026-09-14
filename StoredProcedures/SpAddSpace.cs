using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public enum AddSpaceResult
{
    Success = 0,
    SpaceGroupNotFound = 1
}

public static class SpAddSpace
{
    public const string Name = "sp_AddSpace";

    private const string USER_ID = "@UserId";
    private const string SPACE_GROUP_ID = "@SpaceGroupId";
    private const string SPACE_NAME = "@Name";
    private const string NEW_SPACE_ID = "@NewSpaceId";
    private const string DISPLAY_ORDER = "@DisplayOrder";
    private const string BACKGROUND_IMAGE_UPDATED_AT = "@BackgroundImageUpdatedAt";
    private const string RESULT = "@Result";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {USER_ID} INT,
            {SPACE_GROUP_ID} INT,
            {SPACE_NAME} NVARCHAR(255),
            {NEW_SPACE_ID} INT OUTPUT,
            {DISPLAY_ORDER} INT OUTPUT,
            {BACKGROUND_IMAGE_UPDATED_AT} DATETIME2 OUTPUT,
            {RESULT} INT OUTPUT
        AS
        BEGIN
            SET NOCOUNT ON;
            SET XACT_ABORT ON;
            BEGIN TRANSACTION;

            IF NOT EXISTS (
                SELECT 1
                FROM SpaceGroup sg WITH (UPDLOCK, ROWLOCK)
                INNER JOIN SpaceGroupOwner so ON so.SpaceGroupId = sg.Id
                WHERE sg.Id = {SPACE_GROUP_ID} AND so.UserId = {USER_ID}
            )
            BEGIN
                ROLLBACK TRANSACTION;
                SET {RESULT} = 1;
                RETURN;
            END

            SELECT {DISPLAY_ORDER} = ISNULL(MAX(DisplayOrder), 0) + 1
            FROM Space
            WHERE SpaceGroupId = {SPACE_GROUP_ID};

            INSERT INTO Space (Name, SpaceGroupId, DisplayOrder)
            VALUES ({SPACE_NAME}, {SPACE_GROUP_ID}, {DISPLAY_ORDER});

            SET {NEW_SPACE_ID} = CAST(SCOPE_IDENTITY() AS INT);

            SELECT {BACKGROUND_IMAGE_UPDATED_AT} = BackgroundImageUpdatedAt
            FROM Space
            WHERE Id = {NEW_SPACE_ID};

            SET {RESULT} = 0;

            COMMIT TRANSACTION;
        END
        """;

    public readonly record struct Result(AddSpaceResult Status, int SpaceId, int DisplayOrder, DateTime BackgroundImageUpdatedAt);

    public static async Task<Result> ExecuteAsync(AppDbContext db, int userId, int spaceGroupId, string name, CancellationToken cancellationToken)
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

            var newSpaceIdParam = new SqlParameter(NEW_SPACE_ID, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var displayOrderParam = new SqlParameter(DISPLAY_ORDER, SqlDbType.Int) { Direction = ParameterDirection.Output };
            var backgroundImageUpdatedAtParam = new SqlParameter(BACKGROUND_IMAGE_UPDATED_AT, SqlDbType.DateTime2) { Direction = ParameterDirection.Output };
            var resultParam = new SqlParameter(RESULT, SqlDbType.Int) { Direction = ParameterDirection.Output };

            command.Parameters.Add(new SqlParameter(USER_ID, userId));
            command.Parameters.Add(new SqlParameter(SPACE_GROUP_ID, spaceGroupId));
            command.Parameters.Add(new SqlParameter(SPACE_NAME, name));
            command.Parameters.Add(newSpaceIdParam);
            command.Parameters.Add(displayOrderParam);
            command.Parameters.Add(backgroundImageUpdatedAtParam);
            command.Parameters.Add(resultParam);

            await command.ExecuteNonQueryAsync(cancellationToken);

            var status = (AddSpaceResult)(int)resultParam.Value;

            return status == AddSpaceResult.Success
                ? new Result(status, (int)newSpaceIdParam.Value, (int)displayOrderParam.Value, (DateTime)backgroundImageUpdatedAtParam.Value)
                : new Result(status, 0, 0, default);
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
