using System.ComponentModel.DataAnnotations;
using Acrocuit.Auth;
using Acrocuit.Data;
using Acrocuit.Models;
using Acrocuit.StoredProcedures;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

public class SpaceModel
{
    public int Id { get; set; }

    [Required]
    [StringLength(255, MinimumLength = 1)]
    public required string Name { get; set; }

    [Required]
    public int SpaceGroupId { get; set; }

    public int DisplayOrder { get; set; }
    public DateTime BackgroundImageUpdatedAt { get; set; }
}

public class PositionModel
{
    public int PositionX { get; set; }
    public int PositionY { get; set; }
    public int PositionZ { get; set; }
}

public class DeviceModel
{
    public int Id { get; set; }

    [Required]
    [StringLength(255, MinimumLength = 1)]
    public required string Name { get; set; }
    public int SpaceId { get; set; }
    public PositionModel? Position { get; set; }
}

[ApiController]
[Route("spaces")]
public class SpaceController(AppDbContext db) : ControllerBase
{

    [HttpGet("{spaceId}")]
    public async Task<IActionResult> GetSpace(CurrentUser user, int spaceId, CancellationToken cancellationToken)
    {
        var space = await db.Spaces
            .Where(s => s.Id == spaceId && s.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(s => new SpaceModel()
            {
                Id = s.Id,
                Name = s.Name,
                SpaceGroupId = s.SpaceGroupId,
                DisplayOrder = s.DisplayOrder,
                BackgroundImageUpdatedAt = s.BackgroundImageUpdatedAt
            })
            .FirstOrDefaultAsync(cancellationToken);

        if (space == null)
        {
            return NotFound();
        }

        return Ok(space);
    }

    public class SetSpaceRequest
    {
        [StringLength(255, MinimumLength = 1)]
        public string? Name { get; set; }

        public int? DisplayOrder { get; set; }
    }

    [HttpPut("{spaceId}")]
    public async Task<IActionResult> SetSpace(CurrentUser user, int spaceId, [FromBody] SetSpaceRequest request, CancellationToken cancellationToken)
    {
        var result = await SpSetSpace.ExecuteAsync(db, user.UserId, spaceId, request.Name, request.DisplayOrder, cancellationToken);

        if (result.Status is not SetSpaceResult.Success)
        {
            return NotFound();
        }

        return Ok(new SpaceModel
        {
            Id = spaceId,
            Name = result.Name,
            SpaceGroupId = result.SpaceGroupId,
            DisplayOrder = result.DisplayOrder,
            BackgroundImageUpdatedAt = result.BackgroundImageUpdatedAt
        });
    }

    [HttpDelete("{spaceId}")]
    public async Task<IActionResult> DelSpace(CurrentUser user, int spaceId, CancellationToken cancellationToken)
    {
        var deletedRows = await db.Spaces
            .Where(s => s.Id == spaceId && s.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .ExecuteDeleteAsync(cancellationToken);

        if (deletedRows == 0)
        {
            return NotFound();
        }

        return Ok();
    }


    public class SetSpaceOrderRequest
    {
        [Required]
        public int DisplayOrder { get; set; }
    }

    [HttpPatch("{spaceId}/order")]
    public async Task<IActionResult> SetSpaceOrder(CurrentUser user, int spaceId, SetSpaceOrderRequest request, CancellationToken cancellationToken)
    {
        var result = await SpSetSpace.ExecuteAsync(db, user.UserId, spaceId, null, request.DisplayOrder, cancellationToken);

        if (result.Status is not SetSpaceResult.Success)
        {
            return NotFound();
        }

        return Ok(new SpaceModel
        {
            Id = spaceId,
            Name = result.Name,
            SpaceGroupId = result.SpaceGroupId,
            DisplayOrder = result.DisplayOrder,
            BackgroundImageUpdatedAt = result.BackgroundImageUpdatedAt
        });
    }

    /*
    [HttpPut("{spaceId}/background-image")]
    public async Task<IActionResult> SetSpaceBackgroundImage(CurrentUser user, int spaceId, CancellationToken cancellationToken)
    {

    }
    //*/

    [HttpGet("{spaceId}/breaker-groups")]
    public async Task<IActionResult> GetBreakerGroups(CurrentUser user, int spaceId, CancellationToken cancellationToken)
    {
        var breakerGroups = await db.BreakerGroups
            .Where(b => b.SpaceId == spaceId && b.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(b => new BreakerGroupModel()
            {
                Id = b.Id,
                Name = b.Name,
                SpaceId = b.SpaceId,
                SpaceGroupId = b.SpaceGroupId,
                Position = b.Pos == null ? null : new PositionModel
                {
                    PositionX = b.Pos.PositionX,
                    PositionY = b.Pos.PositionY,
                    PositionZ = b.Pos.PositionZ
                }
            }).ToListAsync(cancellationToken);

        return Ok(breakerGroups);
    }



    [HttpPost("{spaceId}/breaker-groups")]
    public async Task<IActionResult> AddBreakerGroup(CurrentUser user, int spaceId, [FromBody] BreakerGroupModel request, CancellationToken cancellationToken)
    {

        var spaceGroup = await db.Spaces
            .Where(s => s.Id == spaceId && s.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(o => new SpaceGroup()
            {
                Id = o.SpaceGroupId
            }).FirstOrDefaultAsync(cancellationToken);

        if (spaceGroup == null)
        {
            return NotFound();
        }

        var breakerGroup = new BreakerGroup()
        {
            Name = request.Name,
            SpaceId = spaceId,
            SpaceGroupId = spaceGroup.Id,
        };

        if (request.Position is not null)
        {
            breakerGroup.Pos = new BreakerGroupPos
            {
                PositionX = request.Position.PositionX,
                PositionY = request.Position.PositionY,
                PositionZ = request.Position.PositionZ
            };
        }

        db.BreakerGroups.Add(breakerGroup);
        await db.SaveChangesAsync(cancellationToken);

        return Ok(new BreakerGroupModel()
        {
            Id = breakerGroup.Id,
            Name = breakerGroup.Name,
            SpaceId = breakerGroup.SpaceId,
            SpaceGroupId = breakerGroup.SpaceGroupId,
            Position = request.Position
        });
    }

    [HttpGet("{spaceId}/devices")]
    public async Task<IActionResult> GetDevices(CurrentUser user, int spaceId, CancellationToken cancellationToken)
    {
        var devices = await db.Devices
            .Where(d => d.SpaceId == spaceId && d.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(d => new DeviceModel()
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
            }).ToListAsync(cancellationToken);

        return Ok(devices);
    }

    public class CreateDeviceModel
    {
        [Required]
        [StringLength(255, MinimumLength = 1)]
        public required string Name { get; set; }

        public PositionModel? Position { get; set; }
    }

    [HttpPost("{spaceId}/devices")]
    public async Task<IActionResult> AddDevice(CurrentUser user, int spaceId, [FromBody] CreateDeviceModel request, CancellationToken cancellationToken)
    {
        var spaceExists = await db.Spaces
            .AnyAsync(s => s.Id == spaceId && s.SpaceGroup.Owners.Any(o => o.UserId == user.UserId), cancellationToken);

        if (!spaceExists)
        {
            return NotFound();
        }

        var device = new Device
        {
            Name = request.Name,
            SpaceId = spaceId
        };

        if (request.Position is not null)
        {
            device.Pos = new DevicePos
            {
                PositionX = request.Position.PositionX,
                PositionY = request.Position.PositionY,
                PositionZ = request.Position.PositionZ
            };
        }

        db.Devices.Add(device);
        await db.SaveChangesAsync(cancellationToken);

        return Ok(new DeviceModel
        {
            Id = device.Id,
            Name = device.Name,
            SpaceId = device.SpaceId,
            Position = request.Position
        });
    }
}