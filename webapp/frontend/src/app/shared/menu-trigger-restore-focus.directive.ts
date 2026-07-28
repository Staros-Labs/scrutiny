import { Directive, ElementRef, OnDestroy, OnInit, inject } from '@angular/core';
import { MatMenuTrigger } from '@angular/material/menu';
import { Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

// -----------------------------------------------------------------------------------------------------
// Workaround for Starosdev/scrutiny#672
//
// When a mat-menu closes, MatMenuTrigger._destroyMenu() restores focus to the trigger by calling
// focus() without any FocusOptions. The browser then scrolls the trigger back into view, so the page
// jumps back to wherever it was when the menu was opened.
//
// This directive piggybacks on Angular Material's own trigger selector, takes focus restoration over
// from MatMenuTrigger, and restores focus itself with { preventScroll: true }. Keyboard users still
// get focus back on the trigger, but the page no longer moves.
//
// Because it shadows a Material selector, it only applies to components that list it in their
// `imports`. Any new standalone component that renders a [matMenuTriggerFor] must add it too - the
// `Guard menu triggers` step in .github/workflows/ci.yaml enforces that.
// -----------------------------------------------------------------------------------------------------
@Directive({
    // Intentionally matches Angular Material's trigger selector, so the repo prefix rule cannot apply.
    // eslint-disable-next-line @angular-eslint/directive-selector
    selector: '[matMenuTriggerFor]',
})
export class MenuTriggerRestoreFocusDirective implements OnInit, OnDestroy {
    private readonly _menuTrigger = inject(MatMenuTrigger, { self: true });
    private readonly _elementRef = inject<ElementRef<HTMLElement>>(ElementRef);

    // Private
    private _restoreFocus: boolean;
    private readonly _unsubscribeAll: Subject<void>;

    /**
     * Constructor
     */
    constructor() {
        // Set the private defaults
        this._restoreFocus = true;
        this._unsubscribeAll = new Subject();
    }

    // -----------------------------------------------------------------------------------------------------
    // @ Lifecycle hooks
    // -----------------------------------------------------------------------------------------------------

    /**
     * On init
     */
    ngOnInit(): void {
        // Submenu triggers live inside the menu overlay, which is fixed positioned and cannot scroll
        // the page. Material already skips restoring focus to them on non-keyboard closes, so leave
        // that behaviour alone.
        if (this._menuTrigger.triggersSubmenu()) {
            return;
        }

        // Take focus restoration over from Material. The snapshot is refreshed on every open so that
        // a later [matMenuTriggerRestoreFocus] binding write cannot re-arm Material's own focus call.
        this._menuTrigger.menuOpened.pipe(takeUntil(this._unsubscribeAll)).subscribe(() => {
            this._restoreFocus = this._menuTrigger.restoreFocus;
            this._menuTrigger.restoreFocus = false;
        });

        this._menuTrigger.menuClosed.pipe(takeUntil(this._unsubscribeAll)).subscribe(() => {
            // The template opted out of focus restoration
            if (!this._restoreFocus) {
                return;
            }

            // The trigger may already be detached, e.g. a menu item that navigated away via routerLink
            if (!this._elementRef.nativeElement.isConnected) {
                return;
            }

            // Passing no FocusOrigin makes MatMenuTrigger.focus() call the native focus() with our
            // options instead of routing through FocusMonitor.focusVia(), which drops them. The
            // FocusMonitor still derives the origin from the last input modality, so the
            // cdk-mouse-focused / cdk-keyboard-focused classes stay correct.
            this._menuTrigger.focus(undefined, { preventScroll: true });
        });
    }

    /**
     * On destroy
     */
    ngOnDestroy(): void {
        // Unsubscribe from all subscriptions
        this._unsubscribeAll.next();
        this._unsubscribeAll.complete();
    }
}
