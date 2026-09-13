using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class BreakerGroupConfiguration : IEntityTypeConfiguration<BreakerGroup>
{
    public void Configure(EntityTypeBuilder<BreakerGroup> builder)
    {
        builder.ToTable("BreakerGroup");

        builder.HasKey(x => x.Id);
        builder.Property(x => x.Name).HasMaxLength(255).IsRequired();

        builder.HasAlternateKey(x => new { x.Id, x.SpaceGroupId })
            .HasName("UQ_BreakerGroup_Id_SpaceGroupId");

        builder.HasIndex(x => x.SpaceId)
            .HasDatabaseName("IX_BreakerGroup_SpaceId");

        builder.HasOne(x => x.Space)
            .WithMany(x => x.BreakerGroups)
            .HasForeignKey(x => new { x.SpaceId, x.SpaceGroupId })
            .HasPrincipalKey(x => new { x.Id, x.SpaceGroupId })
            .HasConstraintName("FK_BreakerGroup_Space")
            .OnDelete(DeleteBehavior.NoAction);
    }
}
