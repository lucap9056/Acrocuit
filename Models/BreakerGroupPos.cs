namespace Acrocuit.Models;

public class BreakerGroupPos
{
    public int BreakerGroupId { get; set; }
    public int PositionX { get; set; }
    public int PositionY { get; set; }
    public int PositionZ { get; set; }

    public BreakerGroup BreakerGroup { get; set; } = null!;
}
