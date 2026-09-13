namespace Acrocuit.Models;

public class Space
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public int SpaceGroupId { get; set; }
    public int DisplayOrder { get; set; }
    public DateTime BackgroundImageUpdatedAt { get; set; }

    public SpaceGroup SpaceGroup { get; set; } = null!;
    public ICollection<BreakerGroup> BreakerGroups { get; set; } = [];
    public ICollection<Device> Devices { get; set; } = [];
}
