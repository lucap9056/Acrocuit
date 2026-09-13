using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class SpaceConfiguration : IEntityTypeConfiguration<Space>
{
    public void Configure(EntityTypeBuilder<Space> builder)
    {
        builder.ToTable("Space");
        builder.HasKey(x => x.Id);
        builder.Property(x => x.Name).HasMaxLength(255).IsRequired();
        builder.Property(x => x.BackgroundImageUpdatedAt).HasDefaultValueSql("GETUTCDATE()");

        builder.HasAlternateKey(x => new { x.Id, x.SpaceGroupId })
            .HasName("UQ_Space_Id_SpaceGroupId");

        builder.HasIndex(x => new { x.SpaceGroupId, x.DisplayOrder })
            .IsUnique()
            .HasDatabaseName("UQ_SpaceGroup_Order");

        builder.HasIndex(x => x.SpaceGroupId)
            .HasDatabaseName("IX_Space_SpaceGroupId");

        builder.HasOne(x => x.SpaceGroup)
            .WithMany(x => x.Spaces)
            .HasForeignKey(x => x.SpaceGroupId)
            .HasConstraintName("FK_Space_SpaceGroup")
            .OnDelete(DeleteBehavior.NoAction);
    }
}
