using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using RoomBooking.Modules.Identity.Persistence;
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
        return services;
    }
}
