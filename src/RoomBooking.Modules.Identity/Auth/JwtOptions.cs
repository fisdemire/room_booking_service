namespace RoomBooking.Modules.Identity.Auth;

internal sealed class JwtOptions
{
    public const string Section = "Jwt";

    public string Issuer { get; set; } = "room-booking";
    public string Audience { get; set; } = "room-booking-clients";
    public string SigningKey { get; set; } = default!;
    public int ExpiresMinutes { get; set; } = 60;
}
