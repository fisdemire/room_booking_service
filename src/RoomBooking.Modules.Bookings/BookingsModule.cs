using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using RoomBooking.Modules.Bookings.Persistence;
using RoomBooking.SharedKernel.Modules;

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
        return services;
    }
}
