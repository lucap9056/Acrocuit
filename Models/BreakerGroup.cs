namespace Acrocuit.Models;

public class BreakerGroup
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public int SpaceId { get; set; }
    public int SpaceGroupId { get; set; }

    public Space Space { get; set; } = null!;
    public BreakerGroupPos? Pos { get; set; }
    public ICollection<Breaker> Breakers { get; set; } = [];
}
