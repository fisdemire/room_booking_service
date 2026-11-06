using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using RoomBooking.Modules.Rooms.Persistence;
using RoomBooking.SharedKernel.Modules;
using Microsoft.AspNetCore.Routing;
using RoomBooking.Modules.Rooms.Endpoints;

namespace RoomBooking.Modules.Rooms;

public static class RoomsModule
{
    public static IServiceCollection AddRoomsModule(this IServiceCollection services, IConfiguration config)
    {
        services.AddDbContext<RoomsDb>(o => o
            .UseNpgsql(config.GetConnectionString("Default"),
                npgsql => npgsql.MigrationsHistoryTable("__ef_migrations", "rooms"))
            .UseSnakeCaseNamingConvention());

        services.AddSingleton<IModuleInitializer, RoomsInitializer>();
        return services;
    }

    public static IEndpointRouteBuilder MapRoomsEndpoints(this IEndpointRouteBuilder app)
    {
        RoomsEndpoints.Map(app);
        return app;
    }
}
