using System.ComponentModel.DataAnnotations;
using System.Data;
using Acrocuit.Auth;
using Acrocuit.Data;
using Acrocuit.Models;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

public class BreakerGroupModel
{
    public int Id { get; set; }

    [Required]
    [StringLength(255, MinimumLength = 1)]
    public string Name { get; set; } = string.Empty;
    public int SpaceId { get; set; }
    public int SpaceGroupId { get; set; }
    public PositionModel? Position { get; set; }
}

public class BreakerModel
{
    public int Id { get; set; }

    [Required]
    [StringLength(255, MinimumLength = 1)]
    public required string Name { get; set; }
    public int BreakerGroupId { get; set; }
    public int SpaceGroupId { get; set; }
    public int DisplayOrder { get; set; }
    public int? UpstreamBreakerId { get; set; }
}

[ApiController]
[Route("breaker-groups")]
public class BreakerGroupsController(AppDbContext db) : ControllerBase
{
    [HttpGet("{breakerGroupId}")]
    public async Task<IActionResult> GetBreakerGroup(CurrentUser user, int breakerGroupId, CancellationToken cancellationToken)
    {
        var breakerGroup = await db.BreakerGroups
            .Where(b => b.Id == breakerGroupId && b.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(b => new BreakerGroupModel
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
            })
            .FirstOrDefaultAsync(cancellationToken);

        if (breakerGroup is null)
        {
            return NotFound();
        }

        return Ok(breakerGroup);
    }

    public class SetBreakerGroupRequest
    {
        [StringLength(255, MinimumLength = 1)]
        public string? Name { get; set; }

        public PositionModel? Position { get; set; }
    }

    [HttpPut("{breakerGroupId}")]
    public async Task<IActionResult> SetBreakerGroup(CurrentUser user, int breakerGroupId, [FromBody] SetBreakerGroupRequest request, CancellationToken cancellationToken)
    {
        var breakerGroup = await db.BreakerGroups
            .Include(b => b.Pos)
            .Where(b => b.Id == breakerGroupId && b.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .FirstOrDefaultAsync(cancellationToken);

        if (breakerGroup is null)
        {
            return NotFound();
        }

        if (request.Name is not null)
        {
            breakerGroup.Name = request.Name;
        }

        if (request.Position is not null)
        {
            breakerGroup.Pos ??= new BreakerGroupPos();
            breakerGroup.Pos.PositionX = request.Position.PositionX;
            breakerGroup.Pos.PositionY = request.Position.PositionY;
            breakerGroup.Pos.PositionZ = request.Position.PositionZ;
        }

        await db.SaveChangesAsync(cancellationToken);

        return Ok(new BreakerGroupModel
        {
            Id = breakerGroup.Id,
            Name = breakerGroup.Name,
            SpaceId = breakerGroup.SpaceId,
            SpaceGroupId = breakerGroup.SpaceGroupId,
            Position = breakerGroup.Pos is null ? null : new PositionModel
            {
                PositionX = breakerGroup.Pos.PositionX,
                PositionY = breakerGroup.Pos.PositionY,
                PositionZ = breakerGroup.Pos.PositionZ
            }
        });
    }

    public class SetBreakerGroupPositionRequest
    {
        [Required]
        public required PositionModel Position { get; set; }
    }

    [HttpPut("{breakerGroupId}/position")]
    public async Task<IActionResult> SetBreakerGroupPosition(CurrentUser user, int breakerGroupId, [FromBody] SetBreakerGroupPositionRequest request, CancellationToken cancellationToken)
    {
        var breakerGroup = await db.BreakerGroups
            .Include(b => b.Pos)
            .Where(b => b.Id == breakerGroupId && b.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .FirstOrDefaultAsync(cancellationToken);

        if (breakerGroup is null)
        {
            return NotFound();
        }

        breakerGroup.Pos ??= new BreakerGroupPos();
        breakerGroup.Pos.PositionX = request.Position.PositionX;
        breakerGroup.Pos.PositionY = request.Position.PositionY;
        breakerGroup.Pos.PositionZ = request.Position.PositionZ;

        await db.SaveChangesAsync(cancellationToken);

        return Ok(new BreakerGroupModel
        {
            Id = breakerGroup.Id,
            Name = breakerGroup.Name,
            SpaceId = breakerGroup.SpaceId,
            SpaceGroupId = breakerGroup.SpaceGroupId,
            Position = new PositionModel
            {
                PositionX = breakerGroup.Pos.PositionX,
                PositionY = breakerGroup.Pos.PositionY,
                PositionZ = breakerGroup.Pos.PositionZ
            }
        });
    }

    [HttpDelete("{breakerGroupId}")]
    public async Task<IActionResult> DelBreakerGroup(CurrentUser user, int breakerGroupId, CancellationToken cancellationToken)
    {
        var deletedRows = await db.BreakerGroups
            .Where(b => b.Id == breakerGroupId && b.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .ExecuteDeleteAsync(cancellationToken);

        if (deletedRows == 0)
        {
            return NotFound();
        }

        return Ok();
    }

    [HttpGet("{breakerGroupId}/breakers")]
    public async Task<IActionResult> GetBreakers(CurrentUser user, int breakerGroupId, CancellationToken cancellationToken)
    {
        var breakers = await db.Breakers
            .Where(b => b.BreakerGroupId == breakerGroupId && b.BreakerGroup.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .OrderBy(b => b.DisplayOrder)
            .Select(b => new BreakerModel
            {
                Id = b.Id,
                Name = b.Name,
                BreakerGroupId = b.BreakerGroupId,
                SpaceGroupId = b.SpaceGroupId,
                DisplayOrder = b.DisplayOrder,
                UpstreamBreakerId = b.UpstreamBreakerId
            })
            .ToListAsync(cancellationToken);

        return Ok(breakers);
    }

    public class CreateBreakerModel
    {
        [Required]
        [StringLength(255, MinimumLength = 1)]
        public required string Name { get; set; }

        public int? UpstreamBreakerId { get; set; }
    }

    [HttpPost("{breakerGroupId}/breakers")]
    public async Task<IActionResult> AddBreaker(CurrentUser user, int breakerGroupId, [FromBody] CreateBreakerModel request, CancellationToken cancellationToken)
    {
        await using var transaction = await db.Database.BeginTransactionAsync(IsolationLevel.Serializable, cancellationToken);

        var breakerGroup = await db.BreakerGroups
            .Where(b => b.Id == breakerGroupId && b.Space.SpaceGroup.Owners.Any(o => o.UserId == user.UserId))
            .Select(b => new { b.SpaceGroupId })
            .FirstOrDefaultAsync(cancellationToken);

        if (breakerGroup is null)
        {
            return NotFound();
        }

        if (request.UpstreamBreakerId is int upstreamBreakerId)
        {
            var upstreamExists = await db.Breakers
                .AnyAsync(b => b.Id == upstreamBreakerId && b.SpaceGroupId == breakerGroup.SpaceGroupId, cancellationToken);

            if (!upstreamExists)
            {
                return NotFound();
            }
        }

        var maxDisplayOrder = await db.Breakers
            .Where(b => b.BreakerGroupId == breakerGroupId)
            .MaxAsync(b => (int?)b.DisplayOrder, cancellationToken) ?? 0;

        var breaker = new Breaker
        {
            Name = request.Name,
            BreakerGroupId = breakerGroupId,
            SpaceGroupId = breakerGroup.SpaceGroupId,
            DisplayOrder = maxDisplayOrder + 1,
            UpstreamBreakerId = request.UpstreamBreakerId
        };

        db.Breakers.Add(breaker);
        await db.SaveChangesAsync(cancellationToken);
        await transaction.CommitAsync(cancellationToken);

        return Ok(new BreakerModel
        {
            Id = breaker.Id,
            Name = breaker.Name,
            BreakerGroupId = breaker.BreakerGroupId,
            SpaceGroupId = breaker.SpaceGroupId,
            DisplayOrder = breaker.DisplayOrder,
            UpstreamBreakerId = breaker.UpstreamBreakerId
        });
    }
}
