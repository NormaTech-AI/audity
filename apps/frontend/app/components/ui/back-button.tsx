import * as React from "react"
import { buttonVariants, type ButtonProps } from "./button";
import { Slot } from "@radix-ui/react-slot"
import { cn } from "~/lib/utils"
import { useNavigate } from "react-router";
import { ArrowLeft } from 'lucide-react';

const BackButton = React.forwardRef<HTMLButtonElement, ButtonProps>(
    ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    const navigate = useNavigate()
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        onClick={() => navigate(-1)}
        {...props}
      >
        <ArrowLeft className="h-4 w-4" />
      </Comp>
    )
  }
)

export { BackButton, buttonVariants }
