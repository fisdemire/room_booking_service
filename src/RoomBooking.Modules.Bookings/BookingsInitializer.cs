using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using RoomBooking.Modules.Bookings.Persistence;
using RoomBooking.SharedKernel.Modules;

namespace RoomBooking.Modules.Bookings;

internal sealed class BookingsInitializer(IServiceScopeFactory scopes) : IModuleInitializer
{
    public async Task InitializeAsync(CancellationToken ct)
    {
        await using var scope = scopes.CreateAsyncScope();
        await scope.ServiceProvider.GetRequiredService<BookingsDb>().Database.MigrateAsync(ct);
    }
}
