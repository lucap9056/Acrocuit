using System.ComponentModel.DataAnnotations;
using Acrocuit.Auth;
using Acrocuit.Data;
using Acrocuit.Models;
using Acrocuit.StoredProcedures;
using Microsoft.AspNetCore.Mvc;
using Microsoft.EntityFrameworkCore;

public class SpaceGroupModel
{
    public int Id { get; set; }

    [Required]
    [StringLength(255, MinimumLength = 1)]
    public required string Name { get; set; }
}

[ApiController]
[Route("space-groups")]
public class SpaceGroupsController(AppDbContext db) : ControllerBase
{

    [HttpGet]
    public async Task<IActionResult> GetSpaceGroups(CurrentUser user, CancellationToken cancellationToken)
    {
        var spaceGroups = await db.SpaceGroupOwners
            .Where(o => o.UserId == user.UserId)
            .Select(o => new SpaceGroupModel
            {
                Id = o.SpaceGroup.Id,
                Name = o.SpaceGroup.Name
            })
            .ToListAsync(cancellationToken);

        return Ok(spaceGroups);
    }

    [HttpPost]
    public async Task<IActionResult> AddSpaceGroups(CurrentUser user, [FromBody] SpaceGroupModel request, CancellationToken cancellationToken)
    {
        var spaceGroup = new SpaceGroup
        {
            Name = request.Name,
            Owners = [new SpaceGroupOwner { UserId = user.UserId }]
        };

        db.SpaceGroups.Add(spaceGroup);
        await db.SaveChangesAsync(cancellationToken);

        return Ok(new SpaceGroupModel
        {
            Id = spaceGroup.Id,
            Name = spaceGroup.Name
        });
    }

    [HttpGet("{spaceGroupId}")]
    public async Task<IActionResult> GetSpaceGroup(CurrentUser user, int spaceGroupId, CancellationToken cancellationToken)
    {
        var spaceGroup = await db.SpaceGroupOwners
               .Where(o => o.UserId == user.UserId && o.SpaceGroupId == spaceGroupId)
               .Select(o => new SpaceGroupModel
               {
                   Id = o.SpaceGroup.Id,
                   Name = o.SpaceGroup.Name
               })
               .FirstOrDefaultAsync(cancellationToken);

        if (spaceGroup is null)
        {
            return NotFound();
        }

        return Ok(spaceGroup);
    }

    [HttpPut("{spaceGroupId}")]
    public async Task<IActionResult> EditSpaceGroupName(CurrentUser user, int spaceGroupId, [FromBody] SpaceGroupModel request, CancellationToken cancellationToken)
    {
        var updatedRows = await db.SpaceGroupOwners
            .Where(o => o.UserId == user.UserId && o.SpaceGroupId == spaceGroupId)
            .Select(o => o.SpaceGroup)
            .ExecuteUpdateAsync(s => s.SetProperty(b => b.Name, request.Name), cancellationToken);

        if (updatedRows == 0)
        {
            return NotFound();
        }

        return Ok(new SpaceGroupModel
        {
            Id = spaceGroupId,
            Name = request.Name
        });
    }

    [HttpDelete("{spaceGroupId}")]
    public async Task<IActionResult> DelSpaceGroup(CurrentUser user, int spaceGroupId, CancellationToken cancellationToken)
    {
        var deletedRows = await db.SpaceGroups
            .Where(s => s.Id == spaceGroupId && s.Owners.Any(o => o.UserId == user.UserId))
            .ExecuteDeleteAsync(cancellationToken);

        if (deletedRows == 0)
        {
            return NotFound();
        }

        return Ok();
    }

    [HttpGet("{spaceGroupId}/spaces")]
    public async Task<IActionResult> GetSpaces(CurrentUser user, int spaceGroupId, CancellationToken cancellationToken)
    {
        var spaceGroupExists = await db.SpaceGroupOwners
            .AnyAsync(o => o.UserId == user.UserId && o.SpaceGroupId == spaceGroupId, cancellationToken);

        if (!spaceGroupExists)
        {
            return NotFound();
        }

        var spaces = await db.Spaces
            .Where(s => s.SpaceGroupId == spaceGroupId)
            .OrderBy(s => s.DisplayOrder)
            .Select(s => new SpaceModel
            {
                Id = s.Id,
                Name = s.Name,
                SpaceGroupId = s.SpaceGroupId,
                DisplayOrder = s.DisplayOrder,
                BackgroundImageUpdatedAt = s.BackgroundImageUpdatedAt
            })
            .ToListAsync(cancellationToken);

        return Ok(spaces);
    }

    public class CreateSpaceModel
    {
        [Required]
        [StringLength(255, MinimumLength = 1)]
        public required string Name { get; set; }
    }

    [HttpPost("{spaceGroupId}/spaces")]
    public async Task<IActionResult> AddSpace(CurrentUser user, int spaceGroupId, [FromBody] CreateSpaceModel request, CancellationToken cancellationToken)
    {
        var result = await SpAddSpace.ExecuteAsync(db, user.UserId, spaceGroupId, request.Name, cancellationToken);

        if (result.Status is not AddSpaceResult.Success)
        {
            return NotFound();
        }

        return Ok(new SpaceModel
        {
            Id = result.SpaceId,
            Name = request.Name,
            SpaceGroupId = spaceGroupId,
            DisplayOrder = result.DisplayOrder,
            BackgroundImageUpdatedAt = result.BackgroundImageUpdatedAt
        });
    }
}