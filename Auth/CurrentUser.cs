using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using Microsoft.AspNetCore.Mvc;
using Microsoft.AspNetCore.Mvc.ModelBinding;

namespace Acrocuit.Auth;

[ModelBinder(BinderType = typeof(CurrentUserBinder))]
public class CurrentUser
{
    public int UserId { get; set; }

    public static void CurrentUserMiddleware(MvcOptions options)
    {
        options.ModelBinderProviders.Insert(0, new CurrentUserBinderProvider());
    }
}

public class CurrentUserBinderProvider : IModelBinderProvider
{
    public IModelBinder? GetBinder(ModelBinderProviderContext context) =>
        context.Metadata.ModelType == typeof(CurrentUser) ? new CurrentUserBinder() : null;
}

public class CurrentUserBinder : IModelBinder
{
    public Task BindModelAsync(ModelBindingContext bindingContext)
    {
        var value = bindingContext.HttpContext.User.FindFirstValue(JwtRegisteredClaimNames.Sub);

        if (value is null || !int.TryParse(value, out var userId))
        {
            bindingContext.ModelState.TryAddModelError(bindingContext.ModelName, "Current user could not be resolved from the access token.");
            bindingContext.Result = ModelBindingResult.Failed();
            return Task.CompletedTask;
        }

        bindingContext.Result = ModelBindingResult.Success(new CurrentUser { UserId = userId });
        return Task.CompletedTask;
    }
}
