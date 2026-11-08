namespace RoomBooking.Modules.Rooms;

public readonly record struct TimeSlot(DateTimeOffset Start, DateTimeOffset End);

public interface IRoomsModule
{
    Task<bool> ExistsAsync(Guid roomId, CancellationToken ct);
    Task<IReadOnlyList<TimeSlot>> GetSlotsAsync(Guid roomId, DateOnly from, DateOnly to, CancellationToken ct);
}