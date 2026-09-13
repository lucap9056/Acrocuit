using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using Acrocuit.Models;
using Microsoft.IdentityModel.Tokens;

namespace Acrocuit.Auth;

public class JwtTokenService(IConfiguration configuration)
{
    public string GenerateAccessToken(User user) =>
        GenerateToken(user, TokenClaims.AccessTokenUse, TimeSpan.FromMinutes(configuration.GetValue("JwtSettings:AccessTokenExpiryMinutes", 15)));

    public string GenerateRefreshToken(User user) =>
        GenerateToken(user, TokenClaims.RefreshTokenUse, TimeSpan.FromDays(configuration.GetValue("JwtSettings:RefreshTokenExpiryDays", 7)));

    public ClaimsPrincipal? ValidateAccessToken(string token) => ValidateToken(token, TokenClaims.AccessTokenUse);

    public ClaimsPrincipal? ValidateRefreshToken(string token) => ValidateToken(token, TokenClaims.RefreshTokenUse);

    public TokenValidationParameters BuildValidationParameters()
    {
        var jwtSettings = configuration.GetSection("JwtSettings");
        var signingKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtSettings["SecretKey"]!));

        return new TokenValidationParameters
        {
            ValidateIssuer = true,
            ValidIssuer = jwtSettings["Issuer"],

            ValidateAudience = true,
            ValidAudience = jwtSettings["Audience"],

            ValidateIssuerSigningKey = true,
            IssuerSigningKey = signingKey,

            ValidateLifetime = true,
            ClockSkew = TimeSpan.Zero
        };
    }

    private ClaimsPrincipal? ValidateToken(string token, string expectedTokenUse)
    {
        ClaimsPrincipal principal;

        try
        {
            var handler = new JwtSecurityTokenHandler { MapInboundClaims = false };
            principal = handler.ValidateToken(token, BuildValidationParameters(), out _);
        }
        catch (SecurityTokenException)
        {
            return null;
        }

        return principal.FindFirstValue(TokenClaims.TokenUse) == expectedTokenUse ? principal : null;
    }

    private string GenerateToken(User user, string tokenUse, TimeSpan expiry)
    {
        var jwtSettings = configuration.GetSection("JwtSettings");
        var signingKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwtSettings["SecretKey"]!));

        var claims = new[]
        {
            new Claim(JwtRegisteredClaimNames.Sub, user.Id.ToString()),
            new Claim(JwtRegisteredClaimNames.UniqueName, user.Username),
            new Claim(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString()),
            new Claim(TokenClaims.TokenUse, tokenUse)
        };

        var token = new JwtSecurityToken(
            issuer: jwtSettings["Issuer"],
            audience: jwtSettings["Audience"],
            claims: claims,
            expires: DateTime.UtcNow.Add(expiry),
            signingCredentials: new SigningCredentials(signingKey, SecurityAlgorithms.HmacSha256)
        );

        return new JwtSecurityTokenHandler().WriteToken(token);
    }
}
