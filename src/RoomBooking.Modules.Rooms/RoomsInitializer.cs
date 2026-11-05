using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using RoomBooking.Modules.Rooms.Persistence;
using RoomBooking.SharedKernel.Modules;

namespace RoomBooking.Modules.Rooms;

internal sealed class RoomsInitializer(IServiceScopeFactory scopes) : IModuleInitializer
{
    public async Task InitializeAsync(CancellationToken ct)
    {
        await using var scope = scopes.CreateAsyncScope();
        await scope.ServiceProvider.GetRequiredService<RoomsDb>().Database.MigrateAsync(ct);
    }
}
