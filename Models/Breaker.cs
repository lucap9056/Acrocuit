namespace Acrocuit.Models;

public class Breaker
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public int BreakerGroupId { get; set; }
    public int SpaceGroupId { get; set; }
    public int DisplayOrder { get; set; }
    public int? UpstreamBreakerId { get; set; }

    public BreakerGroup BreakerGroup { get; set; } = null!;
    public Breaker? UpstreamBreaker { get; set; }
    public ICollection<Breaker> DownstreamBreakers { get; set; } = [];
    public ICollection<DeviceBreaker> Devices { get; set; } = [];
}
