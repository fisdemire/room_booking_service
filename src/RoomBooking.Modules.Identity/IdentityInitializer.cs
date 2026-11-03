using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.DependencyInjection;
using RoomBooking.Modules.Identity.Persistence;
using RoomBooking.SharedKernel.Modules;

namespace RoomBooking.Modules.Identity;

internal sealed class IdentityInitializer(IServiceScopeFactory scopes) : IModuleInitializer
{
    public async Task InitializeAsync(CancellationToken ct)
    {
        await using var scope = scopes.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<IdentityDb>();
        await db.Database.MigrateAsync(ct);
    }
}
