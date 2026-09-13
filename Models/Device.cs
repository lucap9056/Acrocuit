namespace Acrocuit.Models;

public class Device
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public int SpaceId { get; set; }

    public Space Space { get; set; } = null!;
    public DevicePos? Pos { get; set; }
    public ICollection<DeviceBreaker> Breakers { get; set; } = [];
}
