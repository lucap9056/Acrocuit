using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class DeviceBreakerConfiguration : IEntityTypeConfiguration<DeviceBreaker>
{
    public void Configure(EntityTypeBuilder<DeviceBreaker> builder)
    {
        builder.ToTable("DeviceBreaker");

        builder.HasKey(x => new { x.DeviceId, x.BreakerId })
            .HasName("PK_DeviceBreaker");

        builder.HasIndex(x => x.BreakerId)
            .HasDatabaseName("IX_DeviceBreaker_BreakerId");

        builder.HasOne(x => x.Device)
            .WithMany(x => x.Breakers)
            .HasForeignKey(x => x.DeviceId)
            .HasConstraintName("FK_DeviceBreaker_Device")
            .OnDelete(DeleteBehavior.Cascade);

        builder.HasOne(x => x.Breaker)
            .WithMany(x => x.Devices)
            .HasForeignKey(x => x.BreakerId)
            .HasConstraintName("FK_DeviceBreaker_Breaker")
            .OnDelete(DeleteBehavior.NoAction);
    }
}
