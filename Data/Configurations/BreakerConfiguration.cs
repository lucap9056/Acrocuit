using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class BreakerConfiguration : IEntityTypeConfiguration<Breaker>
{
    public void Configure(EntityTypeBuilder<Breaker> builder)
    {
        builder.ToTable("Breaker", tb =>
            tb.HasCheckConstraint("CHK_Breaker_NotSelfUpstream", "[UpstreamBreaker] IS NULL OR [UpstreamBreaker] <> [Id]"));

        builder.HasKey(x => x.Id);
        builder.Property(x => x.Name).HasMaxLength(255).IsRequired();
        builder.Property(x => x.UpstreamBreakerId).HasColumnName("UpstreamBreaker");

        builder.HasAlternateKey(x => new { x.Id, x.BreakerGroupId })
            .HasName("UQ_Breaker_Id_BreakerGroupId");

        builder.HasIndex(x => new { x.BreakerGroupId, x.DisplayOrder })
            .IsUnique()
            .HasDatabaseName("UQ_BreakerGroup_Order");

        builder.HasIndex(x => x.BreakerGroupId)
            .HasDatabaseName("IX_Breaker_BreakerGroupId");

        builder.HasIndex(x => x.UpstreamBreakerId)
            .HasDatabaseName("IX_Breaker_UpstreamBreaker");

        builder.HasOne(x => x.BreakerGroup)
            .WithMany(x => x.Breakers)
            .HasForeignKey(x => new { x.BreakerGroupId, x.SpaceGroupId })
            .HasPrincipalKey(x => new { x.Id, x.SpaceGroupId })
            .HasConstraintName("FK_Breaker_BreakerGroup")
            .OnDelete(DeleteBehavior.NoAction);

        builder.HasOne(x => x.UpstreamBreaker)
            .WithMany(x => x.DownstreamBreakers)
            .HasForeignKey(x => x.UpstreamBreakerId)
            .HasConstraintName("FK_Breaker_UpstreamBreaker")
            .OnDelete(DeleteBehavior.NoAction);
    }
}
