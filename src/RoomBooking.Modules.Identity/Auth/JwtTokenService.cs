using System.Security.Claims;
using System.Text;
using Microsoft.Extensions.Options;
using Microsoft.IdentityModel.JsonWebTokens;
using Microsoft.IdentityModel.Tokens;
using RoomBooking.Modules.Identity.Domain;

namespace RoomBooking.Modules.Identity.Auth;

internal sealed class JwtTokenService(IOptions<JwtOptions> options)
{
    public (string Token, DateTimeOffset ExpiresAt) Create(User user)
    {
        var o = options.Value;
        var expires = DateTime.UtcNow.AddMinutes(o.ExpiresMinutes);
        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(o.SigningKey));

        var descriptor = new SecurityTokenDescriptor
        {
            Subject = new ClaimsIdentity(new[]
            {
                new Claim("sub", user.Id.ToString()),
                new Claim("email", user.Email),
                new Claim("role", user.Role)
            }),
            Issuer = o.Issuer,
            Audience = o.Audience,
            Expires = expires,
            SigningCredentials = new SigningCredentials(key, SecurityAlgorithms.HmacSha256)
        };

        var token = new JsonWebTokenHandler().CreateToken(descriptor);
        return (token, new DateTimeOffset(expires, TimeSpan.Zero));
    }
}
