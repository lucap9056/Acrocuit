namespace Acrocuit.Models;

public class User
{
    public int Id { get; set; }
    public string Username { get; set; } = string.Empty;

    public UserSecret? Secret { get; set; }
    public ICollection<SpaceGroupOwner> OwnedSpaceGroups { get; set; } = [];
}
