namespace RoomBooking.Modules.Bookings.Domain;

internal static class BookingStatus
{
    public const string Active = "active";
    public const string Cancelled = "cancelled";
}

internal sealed class Booking
{
    public Guid Id { get; set; }
    public Guid RoomId { get; set; }
    public Guid UserId { get; set; }
    public DateTimeOffset StartAt { get; set; }
    public DateTimeOffset EndAt { get; set; }
    public string Status { get; set; } = BookingStatus.Active;
    public string? ConferenceUrl { get; set; }
    public DateTimeOffset CreatedAt { get; set; }
    public DateTimeOffset? CancelledAt { get; set; }
    public DateTimeOffset? ReminderSentAt { get; set; }
}
