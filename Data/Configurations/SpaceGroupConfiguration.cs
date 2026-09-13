using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class SpaceGroupConfiguration : IEntityTypeConfiguration<SpaceGroup>
{
    public void Configure(EntityTypeBuilder<SpaceGroup> builder)
    {
        builder.ToTable("SpaceGroup");
        builder.HasKey(x => x.Id);
        builder.Property(x => x.Name).HasMaxLength(255).IsRequired();
    }
}
