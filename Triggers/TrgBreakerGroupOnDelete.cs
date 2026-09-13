namespace Acrocuit.Triggers;

public static class TrgBreakerGroupOnDelete
{
    public const string Name = "trg_BreakerGroup_OnDelete";

    public const string CreateScript =
        $"""
        CREATE OR ALTER TRIGGER {Name}
        ON BreakerGroup
        INSTEAD OF DELETE
        AS
        BEGIN
            SET NOCOUNT ON;

            DELETE b
            FROM Breaker b
            INNER JOIN deleted d ON b.BreakerGroupId = d.Id;

            DELETE bg
            FROM BreakerGroup bg
            INNER JOIN deleted d ON bg.Id = d.Id;
        END
        """;
}
