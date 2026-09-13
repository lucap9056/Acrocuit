namespace Acrocuit.Triggers;

public static class TrgBreakerOnDelete
{
    public const string Name = "trg_Breaker_OnDelete";

    public const string CreateScript =
        $"""
        CREATE OR ALTER TRIGGER {Name}
        ON Breaker
        INSTEAD OF DELETE
        AS
        BEGIN
            SET NOCOUNT ON;

            UPDATE b
            SET UpstreamBreaker = NULL
            FROM Breaker b
            INNER JOIN deleted d ON b.UpstreamBreaker = d.Id;

            DELETE db
            FROM DeviceBreaker db
            INNER JOIN deleted d ON db.BreakerId = d.Id;

            DELETE b
            FROM Breaker b
            INNER JOIN deleted d ON b.Id = d.Id;
        END
        """;
}
