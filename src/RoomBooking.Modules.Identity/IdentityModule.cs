using System.Text;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Routing;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.IdentityModel.Tokens;
using RoomBooking.Modules.Identity.Auth;
using RoomBooking.Modules.Identity.Domain;
using RoomBooking.Modules.Identity.Endpoints;
using RoomBooking.Modules.Identity.Persistence;
using RoomBooking.SharedKernel.Auth;
using RoomBooking.SharedKernel.Modules;

namespace RoomBooking.Modules.Identity;

public static class IdentityModule
{
    public static IServiceCollection AddIdentityModule(this IServiceCollection services, IConfiguration config)
    {
        services.AddDbContext<IdentityDb>(o => o
            .UseNpgsql(config.GetConnectionString("Default"),
                npgsql => npgsql.MigrationsHistoryTable("__ef_migrations", "identity"))
            .UseSnakeCaseNamingConvention());

        services.AddSingleton<IModuleInitializer, IdentityInitializer>();
        services.AddSingleton<IPasswordHasher<User>, PasswordHasher<User>>();
        services.AddSingleton<JwtTokenService>();

        services.AddOptions<JwtOptions>()
            .Bind(config.GetSection(JwtOptions.Section))
            .Validate(o => !string.IsNullOrEmpty(o.SigningKey) && o.SigningKey.Length >= 32,
                "Jwt:SigningKey должен быть не короче 32 символов.")
            .ValidateOnStart();

        var jwt = config.GetSection(JwtOptions.Section).Get<JwtOptions>()
            ?? throw new InvalidOperationException("Нет секции Jwt в конфигурации.");

        services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
            .AddJwtBearer(o =>
            {
                o.MapInboundClaims = false;
                o.TokenValidationParameters = new TokenValidationParameters
                {
                    ValidateIssuer = true,
                    ValidIssuer = jwt.Issuer,
                    ValidateAudience = true,
                    ValidAudience = jwt.Audience,
                    ValidateIssuerSigningKey = true,
                    IssuerSigningKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(jwt.SigningKey)),
                    ValidateLifetime = true,
                    ClockSkew = TimeSpan.FromSeconds(30),
                    NameClaimType = "sub",
                    RoleClaimType = "role"
                };
            });

        services.AddAuthorizationBuilder()
            .AddPolicy(Policies.AdminOnly, p => p.RequireRole(Roles.Admin));

        return services;
    }

    public static IEndpointRouteBuilder MapIdentityEndpoints(this IEndpointRouteBuilder app)
    {
        AuthEndpoints.Map(app);
        return app;
    }
}
