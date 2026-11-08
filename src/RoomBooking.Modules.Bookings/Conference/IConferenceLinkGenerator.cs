namespace RoomBooking.Modules.Bookings.Conference;

internal interface IConferenceLinkGenerator
{
    Task<string> CreateAsync(Guid bookingId, CancellationToken ct);
}
internal sealed class MockConferenceLinkGenerator : IConferenceLinkGenerator
{
    public Task<string> CreateAsync(Guid bookingId, CancellationToken ct) =>
        Task.FromResult($"https://meet.mock.local/room-booking/{bookingId:N}");
}
