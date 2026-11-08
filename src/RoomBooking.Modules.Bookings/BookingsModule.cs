using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using RoomBooking.Modules.Bookings.Persistence;
using RoomBooking.SharedKernel.Modules;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Routing;
using RoomBooking.Modules.Bookings.Conference;
using RoomBooking.Modules.Bookings.Endpoints;

namespace RoomBooking.Modules.Bookings;

public static class BookingsModule
{
    public static IServiceCollection AddBookingsModule(this IServiceCollection services, IConfiguration config)
    {
        services.AddDbContext<BookingsDb>(o => o
            .UseNpgsql(config.GetConnectionString("Default"),
                npgsql => npgsql.MigrationsHistoryTable("__ef_migrations", "bookings"))
            .UseSnakeCaseNamingConvention());

        services.AddSingleton<IModuleInitializer, BookingsInitializer>();
        services.AddSingleton<IConferenceLinkGenerator, MockConferenceLinkGenerator>();

        return services;
    }

    public static IEndpointRouteBuilder MapBookingsEndpoints(this IEndpointRouteBuilder app)
    {
        BookingsEndpoints.Map(app);
        return app;
    }
}
