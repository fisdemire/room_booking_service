namespace RoomBooking.Modules.Rooms.Domain;

internal sealed class Room
{
    public Guid Id { get; set; }
    public string Name { get; set; } = default!;
    public int Capacity { get; set; }
    public bool IsActive { get; set; } = true;
    public DateTimeOffset CreatedAt { get; set; }
}
