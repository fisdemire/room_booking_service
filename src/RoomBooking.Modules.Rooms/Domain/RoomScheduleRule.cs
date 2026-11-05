namespace RoomBooking.Modules.Rooms.Domain;
internal sealed class RoomScheduleRule
{
    public Guid Id { get; set; }
    public Guid RoomId { get; set; }
    public DayOfWeek DayOfWeek { get; set; }
    public short StartMinute { get; set; }
    public short EndMinute { get; set; }
}
