using RoomBooking.SharedKernel.Auth;

namespace RoomBooking.Modules.Identity.Domain;

internal sealed class User
{
    public Guid Id { get; set; }
    public string Email { get; set; } = default!;
    public string PasswordHash { get; set; } = default!;
    public string Role { get; set; } = Roles.User;
    public DateTimeOffset CreatedAt { get; set; }
}
