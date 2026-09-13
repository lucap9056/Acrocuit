using System.ComponentModel.DataAnnotations;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Security.Cryptography;
using Acrocuit.Auth;
using Acrocuit.Data;
using Acrocuit.Models;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Cryptography.KeyDerivation;
using Microsoft.AspNetCore.Mvc;
using Microsoft.Data.SqlClient;
using Microsoft.EntityFrameworkCore;

[ApiController]
[Route("auth")]
[AllowAnonymous]
public class AuthController(AppDbContext db, JwtTokenService jwtTokenService) : ControllerBase
{
    public class UserCredentials
    {
        [Required]
        [StringLength(127, MinimumLength = 1)]
        public required string Username { get; set; }

        [Required]
        public required string Password { get; set; }
    }

    public class TokenPair
    {
        public string? AccessToken { get; set; }

        [Required]
        public required string RefreshToken { get; set; }
    }

    [HttpPost("register")]
    public async Task<IActionResult> Register([FromBody] UserCredentials userCredentials, CancellationToken cancellationToken)
    {
        byte[] salt = RandomNumberGenerator.GetBytes(32);
        byte[] hash = KeyDerivation.Pbkdf2(
            password: userCredentials.Password,
            salt: salt,
            prf: KeyDerivationPrf.HMACSHA256,
            iterationCount: 100000,
            numBytesRequested: 32
        );

        var user = new User
        {
            Username = userCredentials.Username,
            Secret = new UserSecret
            {
                PasswordSalt = salt,
                PasswordHash = hash
            }
        };

        db.Users.Add(user);

        try
        {
            await db.SaveChangesAsync(cancellationToken);
        }
        catch (DbUpdateException ex) when (ex.InnerException is SqlException { Number: 2601 or 2627 })
        {
            return Conflict(new { message = "Username is already taken. Please choose another one." });
        }

        return Ok(new { message = "Registration successful." });
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] UserCredentials userCredentials, CancellationToken cancellationToken)
    {
        var user = await db.Users
            .Include(u => u.Secret)
            .FirstOrDefaultAsync(u => u.Username == userCredentials.Username, cancellationToken);

        if (user?.Secret is null || !VerifyPassword(userCredentials.Password, user.Secret))
        {
            return Unauthorized(new { message = "Invalid username or password." });
        }

        return Ok(new TokenPair
        {
            AccessToken = jwtTokenService.GenerateAccessToken(user),
            RefreshToken = jwtTokenService.GenerateRefreshToken(user)
        });
    }

    [HttpPost("refresh")]
    public IActionResult Refresh([FromBody] TokenPair request)
    {
        var principal = jwtTokenService.ValidateRefreshToken(request.RefreshToken);

        if (principal is null)
        {
            return Unauthorized(new { message = "Invalid refresh token." });
        }

        var user = new User
        {
            Id = int.Parse(principal.FindFirstValue(JwtRegisteredClaimNames.Sub)!),
            Username = principal.FindFirstValue(JwtRegisteredClaimNames.UniqueName)!
        };

        return Ok(new TokenPair
        {
            AccessToken = jwtTokenService.GenerateAccessToken(user),
            RefreshToken = jwtTokenService.GenerateRefreshToken(user)
        });
    }

    private static bool VerifyPassword(string password, UserSecret secret)
    {
        byte[] hash = KeyDerivation.Pbkdf2(
            password: password,
            salt: secret.PasswordSalt,
            prf: KeyDerivationPrf.HMACSHA256,
            iterationCount: 100000,
            numBytesRequested: 32
        );

        return CryptographicOperations.FixedTimeEquals(hash, secret.PasswordHash);
    }
}