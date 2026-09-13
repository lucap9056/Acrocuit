using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class DevicePosConfiguration : IEntityTypeConfiguration<DevicePos>
{
    public void Configure(EntityTypeBuilder<DevicePos> builder)
    {
        builder.ToTable("DevicePos");
        builder.HasKey(x => x.DeviceId);

        builder.HasOne(x => x.Device)
            .WithOne(x => x.Pos)
            .HasForeignKey<DevicePos>(x => x.DeviceId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
