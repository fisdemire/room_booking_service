using System.Security.Claims;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Routing;
using Microsoft.EntityFrameworkCore;
using RoomBooking.Modules.Bookings.Conference;
using RoomBooking.Modules.Bookings.Domain;
using RoomBooking.Modules.Bookings.Persistence;
using RoomBooking.SharedKernel.Auth;
using RoomBooking.SharedKernel.Database;

namespace RoomBooking.Modules.Bookings.Endpoints;

public sealed record CreateBookingRequest(Guid RoomId, DateTimeOffset StartAt, DateTimeOffset EndAt);

public sealed record BookingDto(
    Guid Id, Guid RoomId, DateTimeOffset StartAt, DateTimeOffset EndAt, string Status, string? ConferenceUrl);

internal static class BookingsEndpoints
{
    public static void Map(IEndpointRouteBuilder app)
    {
        var g = app.MapGroup("/api/bookings").WithTags("Bookings").RequireAuthorization();

        g.MapPost("", Create);
        g.MapGet("/my", My);
        g.MapDelete("/{id:guid}", Cancel);
    }

    private static BookingDto ToDto(Booking b) =>
        new(b.Id, b.RoomId, b.StartAt, b.EndAt, b.Status, b.ConferenceUrl);

    private static async Task<IResult> Create(
        CreateBookingRequest req,
        ClaimsPrincipal user,
        BookingsDb db,
        IConferenceLinkGenerator links,
        CancellationToken ct)
    {
        var booking = new Booking
        {
            Id = Guid.CreateVersion7(),
            RoomId = req.RoomId,
            UserId = user.GetUserId(),
            StartAt = req.StartAt.ToUniversalTime(),
            EndAt = req.EndAt.ToUniversalTime(),
            Status = BookingStatus.Active,
            CreatedAt = DateTimeOffset.UtcNow
        };
        booking.ConferenceUrl = await links.CreateAsync(booking.Id, ct);

        db.Bookings.Add(booking);

        try { await db.SaveChangesAsync(ct); }
        catch (DbUpdateException ex) when (PgErrors.IsKnown(ex)) { return PgErrors.ToProblem(ex); }

        return Results.Created($"/api/bookings/{booking.Id}", ToDto(booking));
    }

    private static async Task<IResult> My(ClaimsPrincipal user, BookingsDb db, CancellationToken ct)
    {
        var uid = user.GetUserId();

        var items = await db.Bookings.AsNoTracking()
            .Where(b => b.UserId == uid)
            .OrderByDescending(b => b.StartAt)
            .ToListAsync(ct);

        return Results.Ok(items.Select(ToDto));
    }

    private static async Task<IResult> Cancel(
        Guid id, ClaimsPrincipal user, BookingsDb db, CancellationToken ct)
    {
        var uid = user.GetUserId();
        var isAdmin = user.IsInRole(Roles.Admin);

        var booking = await db.Bookings.FirstOrDefaultAsync(
            b => b.Id == id && b.Status == BookingStatus.Active && (isAdmin || b.UserId == uid), ct);

        if (booking is null) return Results.NotFound();

        booking.Status = BookingStatus.Cancelled;
        booking.CancelledAt = DateTimeOffset.UtcNow;

        try { await db.SaveChangesAsync(ct); }
        catch (DbUpdateException ex) when (PgErrors.IsKnown(ex)) { return PgErrors.ToProblem(ex); }

        return Results.NoContent();
    }
}
