using Microsoft.EntityFrameworkCore;
using RoomBooking.Modules.Rooms.Persistence;

namespace RoomBooking.Modules.Rooms.Slots;

internal sealed class RoomsModuleApi(RoomsDb db) : IRoomsModule
{
    public Task<bool> ExistsAsync(Guid roomId, CancellationToken ct) =>
        db.Rooms.AnyAsync(r => r.Id == roomId && r.IsActive, ct);

    public async Task<IReadOnlyList<TimeSlot>> GetSlotsAsync(
        Guid roomId, DateOnly from, DateOnly to, CancellationToken ct)
    {
        var rules = await db.ScheduleRules.AsNoTracking()
            .Where(r => r.RoomId == roomId)
            .ToListAsync(ct);

        return SlotGenerator.Generate(rules, from, to).ToList();
    }
}