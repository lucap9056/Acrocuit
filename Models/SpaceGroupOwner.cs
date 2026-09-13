namespace Acrocuit.Models;

public class SpaceGroupOwner
{
    public int UserId { get; set; }
    public int SpaceGroupId { get; set; }

    public User User { get; set; } = null!;
    public SpaceGroup SpaceGroup { get; set; } = null!;
}
