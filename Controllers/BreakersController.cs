using System.ComponentModel.DataAnnotations;
using System.Data;
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
        await using var transaction = await db.Database.BeginTransactionAsync(IsolationLevel.Serializable, cancellationToken);

        var current = await db.Breakers
            .Where(b => b.Id == breakerId && b.BreakerGroup.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(b => new { b.BreakerGroupId, b.DisplayOrder })
            .FirstOrDefaultAsync(cancellationToken);

        if (current is null)
        {
            return NotFound();
        }

        if (request.DisplayOrder is int newOrder && newOrder != current.DisplayOrder)
        {
            await UpdateBreakerOrder(current.BreakerGroupId, breakerId, current.DisplayOrder, newOrder, cancellationToken);
        }

        if (request.Name is not null)
        {
            await db.Breakers
                .Where(b => b.Id == breakerId)
                .ExecuteUpdateAsync(b => b.SetProperty(x => x.Name, request.Name), cancellationToken);
        }

        await transaction.CommitAsync(cancellationToken);

        var breaker = await db.Breakers
            .Where(b => b.Id == breakerId)
            .Select(b => new BreakerModel
            {
                Id = b.Id,
                Name = b.Name,
                BreakerGroupId = b.BreakerGroupId,
                SpaceGroupId = b.SpaceGroupId,
                DisplayOrder = b.DisplayOrder,
                UpstreamBreakerId = b.UpstreamBreakerId
            })
            .FirstAsync(cancellationToken);

        return Ok(breaker);
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
        await using var transaction = await db.Database.BeginTransactionAsync(IsolationLevel.Serializable, cancellationToken);

        var current = await db.Breakers
            .Where(b => b.Id == breakerId && b.BreakerGroup.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(b => new { b.BreakerGroupId, b.DisplayOrder })
            .FirstOrDefaultAsync(cancellationToken);

        if (current is null)
        {
            return NotFound();
        }

        if (request.DisplayOrder != current.DisplayOrder)
        {
            await UpdateBreakerOrder(current.BreakerGroupId, breakerId, current.DisplayOrder, request.DisplayOrder, cancellationToken);
        }

        await transaction.CommitAsync(cancellationToken);

        var breaker = await db.Breakers
            .Where(b => b.Id == breakerId)
            .Select(b => new BreakerModel
            {
                Id = b.Id,
                Name = b.Name,
                BreakerGroupId = b.BreakerGroupId,
                SpaceGroupId = b.SpaceGroupId,
                DisplayOrder = b.DisplayOrder,
                UpstreamBreakerId = b.UpstreamBreakerId
            })
            .FirstAsync(cancellationToken);

        return Ok(breaker);
    }

    public class SetBreakerUpstreamRequest
    {
        public int? UpstreamBreakerId { get; set; }
    }

    [HttpPatch("{breakerId}/upstream")]
    public async Task<IActionResult> SetBreakerUpstream(CurrentUser user, int breakerId, [FromBody] SetBreakerUpstreamRequest request, CancellationToken cancellationToken)
    {
        await using var transaction = await db.Database.BeginTransactionAsync(cancellationToken);

        var result = await SpSetBreakerUpstream.ExecuteAsync(db, user.UserId, breakerId, request.UpstreamBreakerId, cancellationToken);

        if (result is SetBreakerUpstreamResult.SelfReference or SetBreakerUpstreamResult.WouldCreateCycle)
        {
            return BadRequest();
        }

        if (result is not SetBreakerUpstreamResult.Success)
        {
            return NotFound();
        }

        await transaction.CommitAsync(cancellationToken);

        var breaker = await db.Breakers
            .Where(b => b.Id == breakerId)
            .Select(b => new BreakerModel
            {
                Id = b.Id,
                Name = b.Name,
                BreakerGroupId = b.BreakerGroupId,
                SpaceGroupId = b.SpaceGroupId,
                DisplayOrder = b.DisplayOrder,
                UpstreamBreakerId = b.UpstreamBreakerId
            })
            .FirstAsync(cancellationToken);

        return Ok(breaker);
    }

    private async Task UpdateBreakerOrder(int breakerGroupId, int breakerId, int oldOrder, int newOrder, CancellationToken cancellationToken)
    {
        var delta = newOrder - oldOrder;

        if (Math.Abs(delta) == 1)
        {
            await db.Breakers
                .Where(b => b.BreakerGroupId == breakerGroupId && (b.Id == breakerId || b.DisplayOrder == newOrder))
                .ExecuteUpdateAsync(b => b.SetProperty(
                    x => x.DisplayOrder,
                    x => x.Id == breakerId ? newOrder : oldOrder),
                    cancellationToken);
        }
        else if (delta > 0)
        {
            await db.Breakers
                .Where(b => b.BreakerGroupId == breakerGroupId && b.DisplayOrder >= oldOrder && b.DisplayOrder <= newOrder)
                .ExecuteUpdateAsync(b => b.SetProperty(
                    x => x.DisplayOrder,
                    x => x.Id == breakerId ? newOrder : x.DisplayOrder - 1),
                    cancellationToken);
        }
        else
        {
            await db.Breakers
                .Where(b => b.BreakerGroupId == breakerGroupId && b.DisplayOrder >= newOrder && b.DisplayOrder <= oldOrder)
                .ExecuteUpdateAsync(b => b.SetProperty(
                    x => x.DisplayOrder,
                    x => x.Id == breakerId ? newOrder : x.DisplayOrder + 1),
                    cancellationToken);
        }
    }
}
