namespace Acrocuit.Models;

public class SpaceGroup
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;

    public ICollection<Space> Spaces { get; set; } = [];
    public ICollection<SpaceGroupOwner> Owners { get; set; } = [];
}
