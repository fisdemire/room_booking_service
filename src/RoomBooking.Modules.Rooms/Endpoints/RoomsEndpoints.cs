using System.Globalization;
using System.Security.Claims;
using Microsoft.AspNetCore.Builder;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Routing;
using Microsoft.EntityFrameworkCore;
using RoomBooking.Modules.Rooms.Domain;
using RoomBooking.Modules.Rooms.Persistence;
using RoomBooking.SharedKernel.Auth;
using RoomBooking.SharedKernel.Database;

namespace RoomBooking.Modules.Rooms.Endpoints;

public sealed record CreateRoomRequest(string Name, int Capacity);
public sealed record RoomDto(Guid Id, string Name, int Capacity, bool IsActive);
public sealed record AddScheduleRuleRequest(DayOfWeek DayOfWeek, string StartTime, string EndTime);
public sealed record ScheduleRuleDto(Guid Id, DayOfWeek DayOfWeek, string StartTime, string EndTime);

internal static class RoomsEndpoints
{
    public static void Map(IEndpointRouteBuilder app)
    {
        var g = app.MapGroup("/api/rooms").WithTags("Rooms").RequireAuthorization();

        g.MapGet("", async (ClaimsPrincipal user, RoomsDb db, CancellationToken ct) =>
        {
            var query = db.Rooms.AsNoTracking();
            if (!user.IsInRole(Roles.Admin)) query = query.Where(r => r.IsActive);

            var rooms = await query.OrderBy(r => r.Name)
                .Select(r => new RoomDto(r.Id, r.Name, r.Capacity, r.IsActive))
                .ToListAsync(ct);
            return Results.Ok(rooms);
        });

        g.MapPost("", async (CreateRoomRequest req, RoomsDb db, CancellationToken ct) =>
        {
            var room = new Room
            {
                Id = Guid.CreateVersion7(),
                Name = req.Name?.Trim() ?? string.Empty,
                Capacity = req.Capacity,
                IsActive = true,
                CreatedAt = DateTimeOffset.UtcNow
            };
            db.Rooms.Add(room);
            try { await db.SaveChangesAsync(ct); }
            catch (DbUpdateException ex) when (PgErrors.IsKnown(ex)) { return PgErrors.ToProblem(ex); }

            return Results.Created($"/api/rooms/{room.Id}", new RoomDto(room.Id, room.Name, room.Capacity, room.IsActive));
        }).RequireAuthorization(Policies.AdminOnly);

        g.MapPost("/{id:guid}/deactivate", async (Guid id, RoomsDb db, CancellationToken ct) =>
        {
            var rows = await db.Rooms.Where(r => r.Id == id)
                .ExecuteUpdateAsync(s => s.SetProperty(r => r.IsActive, false), ct);
            return rows == 0 ? Results.NotFound() : Results.NoContent();
        }).RequireAuthorization(Policies.AdminOnly);

        g.MapGet("/{id:guid}/schedule", async (Guid id, RoomsDb db, CancellationToken ct) =>
        {
            var rules = await db.ScheduleRules.AsNoTracking()
                .Where(r => r.RoomId == id)
                .OrderBy(r => r.DayOfWeek).ThenBy(r => r.StartMinute)
                .ToListAsync(ct);
            return Results.Ok(rules.Select(r => new ScheduleRuleDto(
                r.Id, r.DayOfWeek, FormatMinute(r.StartMinute), FormatMinute(r.EndMinute))));
        }).RequireAuthorization(Policies.AdminOnly);

        g.MapPost("/{id:guid}/schedule", async (Guid id, AddScheduleRuleRequest req, RoomsDb db, CancellationToken ct) =>
        {
            if (!TryParseMinute(req.StartTime, isEnd: false, out var start) ||
                !TryParseMinute(req.EndTime, isEnd: true, out var end))
                return Results.ValidationProblem(new Dictionary<string, string[]>
                {
                    ["time"] = ["Формат времени HH:mm (конец дня — 00:00)."]
                });

            var rule = new RoomScheduleRule
            {
                Id = Guid.CreateVersion7(),
                RoomId = id,
                DayOfWeek = req.DayOfWeek,
                StartMinute = (short)start,
                EndMinute = (short)end
            };
            db.ScheduleRules.Add(rule);

            try { await db.SaveChangesAsync(ct); }
            catch (DbUpdateException ex) when (PgErrors.IsKnown(ex)) { return PgErrors.ToProblem(ex); }

            return Results.Created($"/api/rooms/{id}/schedule",
                new ScheduleRuleDto(rule.Id, rule.DayOfWeek, FormatMinute(rule.StartMinute), FormatMinute(rule.EndMinute)));
        }).RequireAuthorization(Policies.AdminOnly);
    }

    private static bool TryParseMinute(string? text, bool isEnd, out int minute)
    {
        minute = 0;
        if (!TimeOnly.TryParseExact(text, "HH:mm", CultureInfo.InvariantCulture, DateTimeStyles.None, out var t))
            return false;
        minute = t.Hour * 60 + t.Minute;
        if (isEnd && minute == 0) minute = 1440;
        return true;
    }

    private static string FormatMinute(int minute) =>
        minute == 1440 ? "00:00" : $"{minute / 60:00}:{minute % 60:00}";
}
