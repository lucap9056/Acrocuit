namespace Acrocuit.Triggers;

public static class TrgSpaceOnDelete
{
    public const string Name = "trg_Space_OnDelete";

    public const string CreateScript =
        $"""
        CREATE OR ALTER TRIGGER {Name}
        ON Space
        INSTEAD OF DELETE
        AS
        BEGIN
            SET NOCOUNT ON;

            DELETE bg
            FROM BreakerGroup bg
            INNER JOIN deleted d ON bg.SpaceId = d.Id;

            DELETE s
            FROM Space s
            INNER JOIN deleted d ON s.Id = d.Id;
        END
        """;
}
