using System.ComponentModel.DataAnnotations;
using Acrocuit.Auth;
using Acrocuit.Data;
using Acrocuit.Models;
using Acrocuit.StoredProcedures;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

[ApiController]
[Route("devices")]
public class DevicesController(AppDbContext db) : ControllerBase
{
    [HttpGet("{deviceId}")]
    public async Task<IActionResult> GetDevice(CurrentUser user, int deviceId, CancellationToken cancellationToken)
    {
        var device = await db.Devices
            .Where(d => d.Id == deviceId && d.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(d => new DeviceModel
            {
                Id = d.Id,
                Name = d.Name,
                SpaceId = d.SpaceId,
                Position = d.Pos == null ? null : new PositionModel
                {
                    PositionX = d.Pos.PositionX,
                    PositionY = d.Pos.PositionY,
                    PositionZ = d.Pos.PositionZ
                }
            })
            .FirstOrDefaultAsync(cancellationToken);

        if (device is null)
        {
            return NotFound();
        }

        return Ok(device);
    }

    public class SetDeviceRequest
    {
        [StringLength(255, MinimumLength = 1)]
        public string? Name { get; set; }

        public PositionModel? Position { get; set; }
    }

    [HttpPut("{deviceId}")]
    public async Task<IActionResult> SetDevice(CurrentUser user, int deviceId, [FromBody] SetDeviceRequest request, CancellationToken cancellationToken)
    {
        var device = await db.Devices
            .Include(d => d.Pos)
            .Where(d => d.Id == deviceId && d.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .FirstOrDefaultAsync(cancellationToken);

        if (device is null)
        {
            return NotFound();
        }

        if (request.Name is not null)
        {
            device.Name = request.Name;
        }

        if (request.Position is not null)
        {
            device.Pos ??= new DevicePos();
            device.Pos.PositionX = request.Position.PositionX;
            device.Pos.PositionY = request.Position.PositionY;
            device.Pos.PositionZ = request.Position.PositionZ;
        }

        await db.SaveChangesAsync(cancellationToken);

        return Ok(new DeviceModel
        {
            Id = device.Id,
            Name = device.Name,
            SpaceId = device.SpaceId,
            Position = device.Pos is null ? null : new PositionModel
            {
                PositionX = device.Pos.PositionX,
                PositionY = device.Pos.PositionY,
                PositionZ = device.Pos.PositionZ
            }
        });
    }

    public class SetDevicePositionRequest
    {
        [Required]
        public required PositionModel Position { get; set; }
    }

    [HttpPut("{deviceId}/position")]
    public async Task<IActionResult> SetDevicePosition(CurrentUser user, int deviceId, [FromBody] SetDevicePositionRequest request, CancellationToken cancellationToken)
    {
        var device = await db.Devices
            .Include(d => d.Pos)
            .Where(d => d.Id == deviceId && d.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .FirstOrDefaultAsync(cancellationToken);

        if (device is null)
        {
            return NotFound();
        }

        device.Pos ??= new DevicePos();
        device.Pos.PositionX = request.Position.PositionX;
        device.Pos.PositionY = request.Position.PositionY;
        device.Pos.PositionZ = request.Position.PositionZ;

        await db.SaveChangesAsync(cancellationToken);

        return Ok(new DeviceModel
        {
            Id = device.Id,
            Name = device.Name,
            SpaceId = device.SpaceId,
            Position = new PositionModel
            {
                PositionX = device.Pos.PositionX,
                PositionY = device.Pos.PositionY,
                PositionZ = device.Pos.PositionZ
            }
        });
    }

    [HttpDelete("{deviceId}")]
    public async Task<IActionResult> DelDevice(CurrentUser user, int deviceId, CancellationToken cancellationToken)
    {
        var deletedRows = await db.Devices
            .Where(d => d.Id == deviceId && d.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .ExecuteDeleteAsync(cancellationToken);

        if (deletedRows == 0)
        {
            return NotFound();
        }

        return Ok();
    }

    [HttpGet("{deviceId}/breakers")]
    public async Task<IActionResult> GetDeviceBreakers(CurrentUser user, int deviceId, CancellationToken cancellationToken)
    {
        var deviceExists = await db.Devices
            .AnyAsync(d => d.Id == deviceId && d.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId), cancellationToken);

        if (!deviceExists)
        {
            return NotFound();
        }

        var breakers = await db.DeviceBreakers
            .Where(x => x.DeviceId == deviceId)
            .Select(x => new BreakerModel
            {
                Id = x.Breaker.Id,
                Name = x.Breaker.Name,
                BreakerGroupId = x.Breaker.BreakerGroupId,
                SpaceGroupId = x.Breaker.SpaceGroupId,
                DisplayOrder = x.Breaker.DisplayOrder,
                UpstreamBreakerId = x.Breaker.UpstreamBreakerId
            })
            .ToListAsync(cancellationToken);

        return Ok(breakers);
    }

    public class AddDeviceBreakerRequest
    {
        [Required]
        public int BreakerId { get; set; }
    }

    [HttpPost("{deviceId}/breakers")]
    public async Task<IActionResult> AddDeviceBreaker(CurrentUser user, int deviceId, [FromBody] AddDeviceBreakerRequest request, CancellationToken cancellationToken)
    {
        var result = await SpAddDeviceBreaker.ExecuteAsync(db, user.UserId, deviceId, request.BreakerId, cancellationToken);

        return result switch
        {
            AddDeviceBreakerResult.Success => Ok(),
            AddDeviceBreakerResult.AlreadyLinked => Conflict(),
            _ => NotFound()
        };
    }

    public class DeviceUpstreamModel
    {
        public int BreakerId { get; set; }
        public List<BreakerModel> Chain { get; set; } = [];
    }

    [HttpGet("{deviceId}/upstream")]
    public async Task<IActionResult> GetDeviceUpstream(CurrentUser user, int deviceId, CancellationToken cancellationToken)
    {
        var result = await SpGetDeviceUpstream.ExecuteAsync(db, user.UserId, deviceId, cancellationToken);

        if (!result.Found)
        {
            return NotFound();
        }

        var upstreams = result.Chains.Select(c => new DeviceUpstreamModel
        {
            BreakerId = c.RootBreakerId,
            Chain = [
                .. c.Chain.Select(n => new BreakerModel
                {
                    Id = n.Id,
                    Name = n.Name,
                    BreakerGroupId = n.BreakerGroupId,
                    SpaceGroupId = n.SpaceGroupId,
                    DisplayOrder = n.DisplayOrder,
                    UpstreamBreakerId = n.UpstreamBreakerId
                })
            ]
        }).ToList();

        return Ok(upstreams);
    }

    [HttpDelete("{deviceId}/breakers/{breakerId}")]
    public async Task<IActionResult> DelDeviceBreaker(CurrentUser user, int deviceId, int breakerId, CancellationToken cancellationToken)
    {
        var deletedRows = await db.DeviceBreakers
            .Where(
                x => x.DeviceId == deviceId &&
                x.BreakerId == breakerId &&
                x.Breaker.BreakerGroup.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId)
                )
            .ExecuteDeleteAsync(cancellationToken);

        if (deletedRows == 0)
        {
            return NotFound();
        }

        return Ok();
    }
}
