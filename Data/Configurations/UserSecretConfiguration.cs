using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;
using Microsoft.EntityFrameworkCore.Metadata.Builders;

namespace Acrocuit.Data.Configurations;

public class UserSecretConfiguration : IEntityTypeConfiguration<UserSecret>
{
    public void Configure(EntityTypeBuilder<UserSecret> builder)
    {
        builder.ToTable("UserSecret");

        builder.HasKey(x => x.UserId);
        builder.Property(x => x.PasswordSalt).HasMaxLength(32).IsRequired();
        builder.Property(x => x.PasswordHash).HasMaxLength(64).IsRequired();

        builder.HasOne(x => x.User)
            .WithOne(x => x.Secret)
            .HasForeignKey<UserSecret>(x => x.UserId)
            .OnDelete(DeleteBehavior.Cascade);
    }
}
