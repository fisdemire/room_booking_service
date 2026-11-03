namespace RoomBooking.SharedKernel.Modules;

public interface IModuleInitializer
{
    Task InitializeAsync(CancellationToken ct);
}
