using Acrocuit.Models;
using Microsoft.EntityFrameworkCore;

namespace Acrocuit.Data;

public class AppDbContext(DbContextOptions<AppDbContext> options) : DbContext(options)
{
    public DbSet<SpaceGroup> SpaceGroups => Set<SpaceGroup>();
    public DbSet<Space> Spaces => Set<Space>();
    public DbSet<BreakerGroup> BreakerGroups => Set<BreakerGroup>();
    public DbSet<BreakerGroupPos> BreakerGroupPositions => Set<BreakerGroupPos>();
    public DbSet<Breaker> Breakers => Set<Breaker>();
    public DbSet<Device> Devices => Set<Device>();
    public DbSet<DeviceBreaker> DeviceBreakers => Set<DeviceBreaker>();
    public DbSet<DevicePos> DevicePositions => Set<DevicePos>();
    public DbSet<User> Users => Set<User>();
    public DbSet<UserSecret> UserSecrets => Set<UserSecret>();
    public DbSet<SpaceGroupOwner> SpaceGroupOwners => Set<SpaceGroupOwner>();

    protected override void OnModelCreating(ModelBuilder modelBuilder)
    {
        modelBuilder.ApplyConfigurationsFromAssembly(typeof(AppDbContext).Assembly);
    }
}
