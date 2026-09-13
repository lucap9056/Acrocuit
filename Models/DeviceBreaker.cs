namespace Acrocuit.Models;

public class DeviceBreaker
{
    public int DeviceId { get; set; }
    public int BreakerId { get; set; }

    public Device Device { get; set; } = null!;
    public Breaker Breaker { get; set; } = null!;
}
