import { SignIn } from "@clerk/nextjs"

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="w-full max-w-md">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-foreground">Cashlens</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Sign in to your account
          </p>
        </div>
        <SignIn
          appearance={{
            elements: {
              rootBox: "mx-auto",
              card: "rounded-2xl shadow-md",
              headerTitle: "text-2xl font-bold text-foreground",
              headerSubtitle: "text-muted-foreground",
              socialButtonsBlockButton:
                "rounded-lg border-input hover:bg-accent",
              formButtonPrimary:
                "bg-primary hover:bg-primary/90 text-primary-foreground rounded-lg font-semibold normal-case",
              formFieldLabel: "text-sm font-medium text-foreground",
              formFieldInput:
                "rounded-lg border-input focus:ring-2 focus:ring-ring",
              footerActionLink:
                "text-primary hover:text-primary/90 font-medium",
              identityPreviewText: "text-foreground",
              identityPreviewEditButton: "text-primary hover:text-primary/90",
              formFieldInputShowPasswordButton: "text-muted-foreground",
              otpCodeFieldInput: "rounded-lg border-input",
              formResendCodeLink: "text-primary hover:text-primary/90",
            },
            variables: {
              colorPrimary: "hsl(240, 5.9%, 10%)",
              colorBackground: "hsl(0, 0%, 100%)",
              colorInputBackground: "hsl(0, 0%, 100%)",
              colorInputText: "hsl(240, 10%, 3.9%)",
              borderRadius: "0.5rem",
            },
          }}
        />
      </div>
    </div>
  )
}
