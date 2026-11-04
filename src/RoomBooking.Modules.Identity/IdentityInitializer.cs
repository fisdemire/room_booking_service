using Microsoft.AspNetCore.Identity;
using Microsoft.EntityFrameworkCore;
using Microsoft.Extensions.Configuration;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Logging;
using RoomBooking.Modules.Identity.Domain;
using RoomBooking.Modules.Identity.Persistence;
using RoomBooking.SharedKernel.Auth;
using RoomBooking.SharedKernel.Modules;

namespace RoomBooking.Modules.Identity;

internal sealed class IdentityInitializer(
    IServiceScopeFactory scopes, IConfiguration config, ILogger<IdentityInitializer> log) : IModuleInitializer
{
    public async Task InitializeAsync(CancellationToken ct)
    {
        await using var scope = scopes.CreateAsyncScope();
        var db = scope.ServiceProvider.GetRequiredService<IdentityDb>();
        await db.Database.MigrateAsync(ct);

        var email = config["Seed:Admin:Email"]?.Trim().ToLowerInvariant();
        var password = config["Seed:Admin:Password"];
        if (string.IsNullOrWhiteSpace(email) || string.IsNullOrWhiteSpace(password)) return;
        if (await db.Users.AnyAsync(u => u.Email == email, ct)) return;

        var hasher = scope.ServiceProvider.GetRequiredService<IPasswordHasher<User>>();
        var admin = new User
        {
            Id = Guid.CreateVersion7(),
            Email = email,
            Role = Roles.Admin,
            CreatedAt = DateTimeOffset.UtcNow
        };
        admin.PasswordHash = hasher.HashPassword(admin, password);
        db.Users.Add(admin);
        await db.SaveChangesAsync(ct);
        log.LogInformation("Создан администратор {Email}", email);
    }
}
