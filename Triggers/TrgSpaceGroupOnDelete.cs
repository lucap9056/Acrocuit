namespace Acrocuit.Triggers;

public static class TrgSpaceGroupOnDelete
{
    public const string Name = "trg_SpaceGroup_OnDelete";

    public const string CreateScript =
        $"""
        CREATE OR ALTER TRIGGER {Name}
        ON SpaceGroup
        INSTEAD OF DELETE
        AS
        BEGIN
            SET NOCOUNT ON;

            DELETE s
            FROM Space s
            INNER JOIN deleted d ON s.SpaceGroupId = d.Id;

            DELETE sg
            FROM SpaceGroup sg
            INNER JOIN deleted d ON sg.Id = d.Id;
        END
        """;
}
