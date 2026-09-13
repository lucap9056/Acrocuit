using System.Data;
using Acrocuit.Data;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.StoredProcedures;

public enum AddDeviceBreakerResult
{
    Success = 0,
    DeviceNotFound = 1,
    BreakerNotFound = 2,
    AlreadyLinked = 3
}

public static class SpAddDeviceBreaker
{
    public const string Name = "sp_AddDeviceBreaker";

    private const string USER_ID = "@UserId";
    private const string DEVICE_ID = "@DeviceId";
    private const string BREAKER_ID = "@BreakerId";
    private const string RESULT = "@Result";

    public const string CreateScript =
        $"""
        CREATE OR ALTER PROCEDURE {Name}
            {USER_ID} INT,
            {DEVICE_ID} INT,
            {BREAKER_ID} INT,
            {RESULT} INT OUTPUT
        AS
        BEGIN
            SET NOCOUNT ON;

            DECLARE @SpaceGroupId INT;

            SELECT @SpaceGroupId = s.SpaceGroupId
            FROM Device d
            INNER JOIN Space s ON s.Id = d.SpaceId
            INNER JOIN SpaceGroupOwner so ON so.SpaceGroupId = s.SpaceGroupId
            WHERE d.Id = {DEVICE_ID} AND so.UserId = {USER_ID};

            IF @SpaceGroupId IS NULL
            BEGIN
                SET {RESULT} = 1;
                RETURN;
            END

            IF NOT EXISTS (
                SELECT 1
                FROM Breaker b
                WHERE b.Id = {BREAKER_ID} AND b.SpaceGroupId = @SpaceGroupId
            )
            BEGIN
                SET {RESULT} = 2;
                RETURN;
            END

            IF EXISTS (SELECT 1 FROM DeviceBreaker WHERE DeviceId = {DEVICE_ID} AND BreakerId = {BREAKER_ID})
            BEGIN
                SET {RESULT} = 3;
                RETURN;
            END

            INSERT INTO DeviceBreaker (DeviceId, BreakerId) VALUES ({DEVICE_ID}, {BREAKER_ID});
            SET {RESULT} = 0;
        END
        """;

    public static async Task<AddDeviceBreakerResult> ExecuteAsync(AppDbContext db, int userId, int deviceId, int breakerId, CancellationToken cancellationToken)
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
            command.Parameters.Add(new SqlParameter(DEVICE_ID, deviceId));
            command.Parameters.Add(new SqlParameter(BREAKER_ID, breakerId));
            command.Parameters.Add(resultParam);

            await command.ExecuteNonQueryAsync(cancellationToken);

            return (AddDeviceBreakerResult)(int)resultParam.Value;
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
