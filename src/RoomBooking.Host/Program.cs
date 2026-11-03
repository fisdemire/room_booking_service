using RoomBooking.Modules.Identity;
using RoomBooking.SharedKernel.Modules;

var builder = WebApplication.CreateBuilder(args);

builder.Services.AddIdentityModule(builder.Configuration);

var app = builder.Build();

foreach (var initializer in app.Services.GetServices<IModuleInitializer>())
    await initializer.InitializeAsync(app.Lifetime.ApplicationStopping);

app.MapGet("/health", () => Results.Ok(new { status = "ok" }));

app.Run();
