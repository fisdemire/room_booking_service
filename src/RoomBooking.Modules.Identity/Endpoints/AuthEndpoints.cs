using System.Net.Mail;
using System.Security.Claims;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Identity;
using Microsoft.AspNetCore.Routing;
using Microsoft.EntityFrameworkCore;
using RoomBooking.Modules.Identity.Auth;
using RoomBooking.Modules.Identity.Domain;
using RoomBooking.Modules.Identity.Persistence;
using RoomBooking.SharedKernel.Auth;
using RoomBooking.SharedKernel.Database;

namespace RoomBooking.Modules.Identity.Endpoints;

public sealed record RegisterRequest(string Email, string Password);
public sealed record LoginRequest(string Email, string Password);
public sealed record LoginResponse(string Token, DateTimeOffset ExpiresAt, string Role);

internal static class AuthEndpoints
{
    public static void Map(IEndpointRouteBuilder app)
    {
        var g = app.MapGroup("/api/auth").WithTags("Auth");

        g.MapPost("/register", async (RegisterRequest req, IdentityDb db,
            IPasswordHasher<User> hasher, CancellationToken ct) =>
        {
            var email = (req.Email ?? string.Empty).Trim().ToLowerInvariant();
            if (!MailAddress.TryCreate(email, out _) || (req.Password?.Length ?? 0) < 8)
                return Results.ValidationProblem(new Dictionary<string, string[]>
                {
                    ["request"] = ["Некорректный email или пароль короче 8 символов."]
                });

            var user = new User
            {
                Id = Guid.CreateVersion7(),
                Email = email,
                Role = Roles.User,
                CreatedAt = DateTimeOffset.UtcNow
            };
            user.PasswordHash = hasher.HashPassword(user, req.Password!);

            db.Users.Add(user);
            try { await db.SaveChangesAsync(ct); }
            catch (DbUpdateException ex) when (PgErrors.IsKnown(ex)) { return PgErrors.ToProblem(ex); }

            return Results.Created($"/api/auth/users/{user.Id}", new { user.Id, user.Email, user.Role });
        });

        g.MapPost("/login", async (LoginRequest req, IdentityDb db,
            IPasswordHasher<User> hasher, JwtTokenService jwt, CancellationToken ct) =>
        {
            var email = (req.Email ?? string.Empty).Trim().ToLowerInvariant();
            var user = await db.Users.AsNoTracking().SingleOrDefaultAsync(u => u.Email == email, ct);

            if (user is null ||
                hasher.VerifyHashedPassword(user, user.PasswordHash, req.Password ?? string.Empty)
                    == PasswordVerificationResult.Failed)
                return Results.Unauthorized();

            var (token, expiresAt) = jwt.Create(user);
            return Results.Ok(new LoginResponse(token, expiresAt, user.Role));
        });

        g.MapGet("/me", (ClaimsPrincipal principal) => Results.Ok(new
        {
            id = principal.GetUserId(),
            email = principal.FindFirstValue("email"),
            role = principal.FindFirstValue("role")
        })).RequireAuthorization();
    }
}
