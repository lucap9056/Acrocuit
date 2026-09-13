using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class DeviceConfiguration : IEntityTypeConfiguration<Device>
{
    public void Configure(EntityTypeBuilder<Device> builder)
    {
        builder.ToTable("Device");

        builder.HasKey(x => x.Id);
        builder.Property(x => x.Name).HasMaxLength(255).IsRequired();

        builder.HasIndex(x => x.SpaceId)
            .HasDatabaseName("IX_Device_SpaceId");

        builder.HasOne(x => x.Space)
            .WithMany(x => x.Devices)
            .HasForeignKey(x => x.SpaceId)
            .HasConstraintName("FK_Device_Space")
            .OnDelete(DeleteBehavior.Cascade);
    }
}
