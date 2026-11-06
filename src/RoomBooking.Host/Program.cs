using RoomBooking.Modules.Identity;
using RoomBooking.SharedKernel.Modules;
using RoomBooking.Modules.Rooms;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddIdentityModule(builder.Configuration);
builder.Services.AddRoomsModule(builder.Configuration);

var app = builder.Build();

foreach (var initializer in app.Services.GetServices<IModuleInitializer>())
    await initializer.InitializeAsync(app.Lifetime.ApplicationStopping);

app.UseAuthentication();
app.UseAuthorization();

app.MapGet("/health", () => Results.Ok(new { status = "ok" }));
app.MapIdentityEndpoints();
app.MapRoomsEndpoints();

app.Run();

