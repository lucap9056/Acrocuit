namespace Acrocuit.Models;

public class DevicePos
{
    public int DeviceId { get; set; }
    public int PositionX { get; set; }
    public int PositionY { get; set; }
    public int PositionZ { get; set; }

    public Device Device { get; set; } = null!;
}
