using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class SpaceGroupOwnerConfiguration : IEntityTypeConfiguration<SpaceGroupOwner>
{
    public void Configure(EntityTypeBuilder<SpaceGroupOwner> builder)
    {
        builder.ToTable("SpaceGroupOwner");
        builder.HasKey(x => new { x.UserId, x.SpaceGroupId })
            .HasName("PK_SpaceGroupOwner");

        builder.HasOne(x => x.User)
            .WithMany(x => x.OwnedSpaceGroups)
            .HasForeignKey(x => x.UserId)
            .HasConstraintName("FK_SpaceGroupOwner_User")
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasOne(x => x.SpaceGroup)
            .WithMany(x => x.Owners)
            .HasForeignKey(x => x.SpaceGroupId)
            .HasConstraintName("FK_SpaceGroupOwner_SpaceGroup")
            .OnDelete(DeleteBehavior.Cascade);
    }
}
