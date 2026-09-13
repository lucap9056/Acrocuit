using System.Security.Claims;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Acrocuit.Auth;
using Acrocuit.Data;
using Acrocuit.StoredProcedures;
using Acrocuit.Triggers;
using Microsoft.EntityFrameworkCore;
using Microsoft.AspNetCore.Authorization;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddEndpointsApiExplorer();
builder.Services.AddSwaggerGen();
builder.Services.AddProblemDetails();

builder.Services.AddDbContext<AppDbContext>(options =>
    options.UseSqlServer(builder.Configuration.GetConnectionString("Default")));

var jwtTokenService = new JwtTokenService(builder.Configuration);
builder.Services.AddSingleton(jwtTokenService);

builder.Services.AddAuthentication(options =>
{
    options.DefaultAuthenticateScheme = JwtBearerDefaults.AuthenticationScheme;
    options.DefaultChallengeScheme = JwtBearerDefaults.AuthenticationScheme;
})
.AddJwtBearer(options =>
{
    options.MapInboundClaims = false;
    options.TokenValidationParameters = jwtTokenService.BuildValidationParameters();

    options.Events = new JwtBearerEvents
    {
        OnTokenValidated = context =>
        {
            if (context.Principal?.FindFirstValue(TokenClaims.TokenUse) != TokenClaims.AccessTokenUse)
            {
                context.Fail("Refresh tokens cannot be used to access protected resources.");
            }

            return Task.CompletedTask;
        }
    };
});

builder.Services.AddAuthorizationBuilder()
    .SetFallbackPolicy(
        new AuthorizationPolicyBuilder().RequireAuthenticatedUser().Build()
    );

builder.Services.AddControllers(CurrentUser.CurrentUserMiddleware);

var app = builder.Build();

using (var scope = app.Services.CreateScope())
{
    var db = scope.ServiceProvider.GetRequiredService<AppDbContext>();
    string[] storedProcedureScripts =
    [
        SpGetBreakerUpstreamChain.CreateScript,
        SpGetBreakerDownstream.CreateScript,
        SpAddDeviceBreaker.CreateScript,
        SpSetBreakerUpstream.CreateScript
    ];

    foreach (var script in storedProcedureScripts)
    {
        await db.Database.ExecuteSqlRawAsync(script);
    }

    string[] triggerScripts =
    [
        TrgBreakerOnDelete.CreateScript,
        TrgBreakerGroupOnDelete.CreateScript,
        TrgSpaceOnDelete.CreateScript,
        TrgSpaceGroupOnDelete.CreateScript
    ];

    foreach (var script in triggerScripts)
    {
        await db.Database.ExecuteSqlRawAsync(script);
    }
}

if (app.Environment.IsDevelopment())
{
    app.UseDeveloperExceptionPage();
    app.UseSwagger();
    app.UseSwaggerUI();
}
else
{
    app.UseExceptionHandler();
}

app.UseAuthentication();
app.UseAuthorization();

app.UseHttpsRedirection();

app.MapControllers();

await app.RunAsync();