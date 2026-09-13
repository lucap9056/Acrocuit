using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class BreakerGroupPosConfiguration : IEntityTypeConfiguration<BreakerGroupPos>
{
    public void Configure(EntityTypeBuilder<BreakerGroupPos> builder)
    {
        builder.ToTable("BreakerGroupPos");
        builder.HasKey(x => x.BreakerGroupId);

        builder.HasOne(x => x.BreakerGroup)
            .WithOne(x => x.Pos)
            .HasForeignKey<BreakerGroupPos>(x => x.BreakerGroupId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
