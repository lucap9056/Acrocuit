using System.ComponentModel.DataAnnotations;
using Acrocuit.Auth;
using Acrocuit.Data;
using Acrocuit.Models;
using Acrocuit.StoredProcedures;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

public class DeviceTreeNodeModel
{
    public int Id { get; set; }
    public required string Name { get; set; }
    public int SpaceId { get; set; }
    public int BreakerId { get; set; }
}

public class BreakerDownstreamModel
{
    public List<BreakerModel> Breakers { get; set; } = [];
    public List<DeviceTreeNodeModel> Devices { get; set; } = [];
}

[ApiController]
[Route("breakers")]
public class BreakersController(AppDbContext db) : ControllerBase
{
    [HttpGet("{breakerId}/downstream")]
    public async Task<IActionResult> GetBreakerDownstream(CurrentUser user, int breakerId, CancellationToken cancellationToken)
    {
        var result = await SpGetBreakerDownstream.ExecuteAsync(db, breakerId, user.UserId, cancellationToken);

        if (result.Breakers.Count == 0)
        {
            return NotFound();
        }

        return Ok(new BreakerDownstreamModel
        {
            Breakers = [
                .. result.Breakers.Select(n => new BreakerModel
                {
                    Id = n.Id,
                    Name = n.Name,
                    BreakerGroupId = n.BreakerGroupId,
                    SpaceGroupId = n.SpaceGroupId,
                    DisplayOrder = n.DisplayOrder,
                    UpstreamBreakerId = n.UpstreamBreakerId
                })
            ],
            Devices = [
                .. result.Devices.Select(d => new DeviceTreeNodeModel
                {
                    Id = d.Id,
                    Name = d.Name,
                    SpaceId = d.SpaceId,
                    BreakerId = d.BreakerId
                })
            ]
        });
    }

    [HttpGet("{breakerId}")]
    public async Task<IActionResult> GetBreaker(CurrentUser user, int breakerId, CancellationToken cancellationToken)
    {
        var breaker = await db.Breakers
            .Where(b => b.Id == breakerId && b.BreakerGroup.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(b => new BreakerModel
            {
                Id = b.Id,
                Name = b.Name,
                BreakerGroupId = b.BreakerGroupId,
                SpaceGroupId = b.SpaceGroupId,
                DisplayOrder = b.DisplayOrder,
                UpstreamBreakerId = b.UpstreamBreakerId
            })
            .FirstOrDefaultAsync(cancellationToken);

        if (breaker is null)
        {
            return NotFound();
        }

        return Ok(breaker);
    }

    public class SetBreakerRequest
    {
        [StringLength(255, MinimumLength = 1)]
        public string? Name { get; set; }

        public int? DisplayOrder { get; set; }
    }

    [HttpPut("{breakerId}")]
    public async Task<IActionResult> SetBreaker(CurrentUser user, int breakerId, [FromBody] SetBreakerRequest request, CancellationToken cancellationToken)
    {
        var result = await SpSetBreaker.ExecuteAsync(db, user.UserId, breakerId, request.Name, request.DisplayOrder, cancellationToken);

        if (result.Status is not SetBreakerResult.Success)
        {
            return NotFound();
        }

        return Ok(new BreakerModel
        {
            Id = breakerId,
            Name = result.Name,
            BreakerGroupId = result.BreakerGroupId,
            SpaceGroupId = result.SpaceGroupId,
            DisplayOrder = result.DisplayOrder,
            UpstreamBreakerId = result.UpstreamBreakerId
        });
    }

    [HttpDelete("{breakerId}")]
    public async Task<IActionResult> DelBreaker(CurrentUser user, int breakerId, CancellationToken cancellationToken)
    {
        var deletedRows = await db.Breakers
            .Where(b => b.Id == breakerId && b.BreakerGroup.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .ExecuteDeleteAsync(cancellationToken);

        if (deletedRows == 0)
        {
            return NotFound();
        }

        return Ok();
    }

    public class SetBreakerOrderRequest
    {
        [Required]
        public int DisplayOrder { get; set; }
    }

    [HttpPatch("{breakerId}/order")]
    public async Task<IActionResult> SetBreakerOrder(CurrentUser user, int breakerId, [FromBody] SetBreakerOrderRequest request, CancellationToken cancellationToken)
    {
        var result = await SpSetBreaker.ExecuteAsync(db, user.UserId, breakerId, null, request.DisplayOrder, cancellationToken);

        if (result.Status is not SetBreakerResult.Success)
        {
            return NotFound();
        }

        return Ok(new BreakerModel
        {
            Id = breakerId,
            Name = result.Name,
            BreakerGroupId = result.BreakerGroupId,
            SpaceGroupId = result.SpaceGroupId,
            DisplayOrder = result.DisplayOrder,
            UpstreamBreakerId = result.UpstreamBreakerId
        });
    }

    public class SetBreakerUpstreamRequest
    {
        public int? UpstreamBreakerId { get; set; }
    }

    [HttpPatch("{breakerId}/upstream")]
    public async Task<IActionResult> SetBreakerUpstream(CurrentUser user, int breakerId, [FromBody] SetBreakerUpstreamRequest request, CancellationToken cancellationToken)
    {
        var result = await SpSetBreakerUpstream.ExecuteAsync(db, user.UserId, breakerId, request.UpstreamBreakerId, cancellationToken);

        if (result.Status is SetBreakerUpstreamResult.SelfReference or SetBreakerUpstreamResult.WouldCreateCycle)
        {
            return BadRequest();
        }

        if (result.Status is not SetBreakerUpstreamResult.Success)
        {
            return NotFound();
        }

        return Ok(new BreakerModel
        {
            Id = breakerId,
            Name = result.Name,
            BreakerGroupId = result.BreakerGroupId,
            SpaceGroupId = result.SpaceGroupId,
            DisplayOrder = result.DisplayOrder,
            UpstreamBreakerId = result.UpstreamBreakerId
        });
    }

}
