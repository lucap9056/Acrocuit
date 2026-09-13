namespace Acrocuit.Models;

public class UserSecret
{
    public int UserId { get; set; }
    public byte[] PasswordSalt { get; set; } = [];
    public byte[] PasswordHash { get; set; } = [];

    public User User { get; set; } = null!;
}
