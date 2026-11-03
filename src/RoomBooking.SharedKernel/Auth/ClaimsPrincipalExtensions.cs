using System.Security.Claims;

namespace RoomBooking.SharedKernel.Auth;

public static class ClaimsPrincipalExtensions
{
    public static Guid GetUserId(this ClaimsPrincipal principal)
    {
        var value = principal.FindFirstValue("sub")
            ?? throw new InvalidOperationException("В токене нет claim 'sub'.");
        return Guid.Parse(value);
    }
}
